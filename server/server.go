package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/golang/gddo/httputil"
	"github.com/gorilla/mux"
	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/metadata/content"
	"github.com/rancher/metadata/content/memory"
	"github.com/rancher/metadata/subscriber"
	"github.com/rancher/metadata/types"
)

const (
	ContentText = 1
	ContentJSON = 2
)

// Server specifies the configuration for the metadata server
type Server struct {
	listen     string
	enableXff  bool
	subscriber *subscriber.Subscriber
	store      content.Store
}

func New(opts *client.ClientOpts, listen string, enableXff bool) (*Server, error) {
	s := &Server{
		listen:    listen,
		enableXff: enableXff,
		store:     memory.NewMemoryStore(context.Background()),
	}

	subscriber, err := subscriber.NewSubscriber(opts, s.store)
	if err != nil {
		return nil, err
	}

	s.subscriber = subscriber
	return s, nil
}

func (s *Server) Start() error {
	go s.runServer()
	s.subscriber.Start()
	return fmt.Errorf("Server died")
}

func (s *Server) runServer() {
	s.watchSignals()

	router := mux.NewRouter()

	router.HandleFunc("/", s.root).
		Methods("GET", "HEAD").
		Name("Root")

	router.HandleFunc("/{version}", s.metadata).
		Methods("GET", "HEAD").
		Name("Version")

	router.HandleFunc("/{version}/{key:.*}", s.metadata).
		Queries("wait", "true", "value", "{oldValue}").
		Methods("GET", "HEAD").
		Name("Wait")

	router.HandleFunc("/{version}/{key:.*}", s.metadata).
		Methods("GET", "HEAD").
		Name("Metadata")

	logrus.Info("Listening on ", s.listen)
	logrus.Fatal(http.ListenAndServe(s.listen, router))
}

func (s *Server) lookupAnswer(wait bool, oldValue, version string, ip string, path []string, maxWait time.Duration) (interface{}, bool) {
	if maxWait == time.Duration(0) {
		maxWait = 10 * time.Second
	}

	if maxWait > 2*time.Minute {
		maxWait = 2 * time.Minute
	}

	start := time.Now()

	for {
		val, ok := s.getValue(version, ip, path)
		if !wait {
			return val, ok
		}
		if time.Now().Sub(start) > maxWait {
			return val, ok
		}
		if ok && fmt.Sprint(val) != oldValue {
			return val, ok
		}

		s.store.WaitChanged()
	}
}

func (s *Server) getValue(version, ip string, path []string) (interface{}, bool) {
	var root interface{}

	if len(path) > 0 && path[0] == "self" {
		root = content.NewSelfObject(version, ip, s.store)
		path = path[1:]
	} else {
		env, ok := content.GetEnvironment(s.store, version, ip)
		if !ok {
			return nil, false
		}

		root = env
	}

	if len(path) == 0 {
		return root, true
	}

	return traverse(root, path)
}

func getIndexed(value reflect.Value, index string) (interface{}, bool) {
	idx, err := strconv.Atoi(index)
	if err == nil {
		if idx < value.Len() {
			return value.Index(idx).Interface(), true
		}
		return nil, false
	}

	for i := 0; i < value.Len(); i++ {
		obj := value.Index(i).Interface()
		if named, ok := obj.(interface {
			Name() string
		}); ok {
			if strings.EqualFold(named.Name(), index) {
				return obj, true
			}
		}
	}

	return nil, false
}

func traverse(in interface{}, path []string) (interface{}, bool) {
	out := in

	for _, key := range path {
		valid := false

		switch v := out.(type) {
		case types.Object:
			out, valid = v.Get(key)
			if !valid {
				out, valid = v.Get(strings.ToLower(key))
			}
		case map[string]interface{}:
			out, valid = v[key]
			if !valid {
				out, valid = v[strings.ToLower(key)]
			}
		default:
			if reflect.TypeOf(v).Kind() == reflect.Slice {
				out, valid = getIndexed(reflect.ValueOf(v), key)
			} else {
				logrus.Debug("Unknown type %T at /%s", v, path)
			}
		}

		if !valid {
			return nil, false
		}
	}

	return out, true
}

func (s *Server) watchSignals() {
	reloadChan := make(chan os.Signal, 1)
	signal.Notify(reloadChan, syscall.SIGHUP)

	go func() {
		for range reloadChan {
			s.subscriber.Reload()
		}
	}()
}

func contentType(req *http.Request) int {
	str := httputil.NegotiateContentType(req, []string{
		"text/plain",
		"application/json",
	}, "text/plain")

	if strings.Contains(str, "json") {
		return ContentJSON
	}

	return ContentText
}

func (s *Server) root(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	logrus.WithFields(logrus.Fields{
		"client":  s.requestIP(req),
		"version": "root",
	}).Debugf("OK: %s", "/")

	// This will always succeed, don't need to check ok
	m, _ := content.GetEnvironment(s.store, "/", "")
	respondSuccess(w, req, m)
}

func (s *Server) metadata(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(req)
	clientIP := s.requestIP(req)

	version := vars["version"]
	wait := mux.CurrentRoute(req).GetName() == "Wait"
	oldValue := vars["oldValue"]
	maxWait, _ := strconv.Atoi(req.URL.Query().Get("maxWait"))
	path := strings.TrimRight(req.URL.EscapedPath()[1:], "/")
	pathSegments := strings.Split(path, "/")[1:]
	displayKey := ""
	var err error
	for i := 0; err == nil && i < len(pathSegments); i++ {
		displayKey += "/" + pathSegments[i]
		pathSegments[i], err = url.QueryUnescape(pathSegments[i])
	}

	if err != nil {
		respondError(w, req, err.Error(), http.StatusBadRequest)
		return
	}

	logrus.WithFields(logrus.Fields{
		"version":  version,
		"client":   clientIP,
		"wait":     wait,
		"oldValue": oldValue,
		"maxWait":  maxWait}).Debugf("Searching for: %s", displayKey)
	val, ok := s.lookupAnswer(wait, oldValue, version, clientIP, pathSegments, time.Duration(maxWait)*time.Second)

	if ok {
		logrus.WithFields(logrus.Fields{
			"version": version,
			"client":  clientIP,
		}).Debugf("OK: %s", displayKey)
		respondSuccess(w, req, val)
	} else {
		logrus.WithFields(logrus.Fields{
			"version": version,
			"client":  clientIP,
		}).Infof("Error: %s", displayKey)
		respondError(w, req, "Not found", http.StatusNotFound)
	}
}

func (s *Server) requestIP(req *http.Request) string {
	if s.enableXff {
		clientIP := req.Header.Get("X-Forwarded-For")
		if len(clientIP) > 0 {
			return clientIP
		}
	}

	clientIP, _, _ := net.SplitHostPort(req.RemoteAddr)
	return clientIP
}
