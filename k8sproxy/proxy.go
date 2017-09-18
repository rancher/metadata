package k8sproxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"context"

	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/websocket-proxy/proxy/websocket"
	"golang.org/x/sync/errgroup"
)

type Proxy struct {
	c               *client.RancherClient
	prefix          string
	cattleAddr      string
	scheme          string
	reverseProxy    *httputil.ReverseProxy
	httpListenAddr  string
	httpsListenAddr string
	certFile        string
	keyFile         string
}

func New(c *client.RancherClient, httpListenAddr, httpsListenAddr, certFile, keyFile string) (*Proxy, error) {
	agents, err := c.Agent.List(nil)
	if err != nil {
		return nil, err
	}

	if len(agents.Data) != 1 {
		return nil, fmt.Errorf("Failed to find 1 agent, found %d", len(agents.Data))
	}

	if agents.Data[0].ClusterId == "" {
		return nil, fmt.Errorf("Failed to find cluster ID, cluster ID is not set: %#v", agents.Data[0])
	}

	cattleURL, err := url.Parse(c.GetOpts().Url)
	if err != nil {
		return nil, errors.Wrap(err, "Looking up cattle address")
	}
	cattleURL.Path = "/"

	logrus.Infof("Running proxy for cluster ID %s", agents.Data[0].ClusterId)

	reverseProxy := httputil.NewSingleHostReverseProxy(cattleURL)
	reverseProxy.FlushInterval = 100 * time.Millisecond

	return &Proxy{
		c:               c,
		prefix:          "/k8s/clusters/" + agents.Data[0].ClusterId,
		cattleAddr:      cattleURL.Host,
		reverseProxy:    reverseProxy,
		scheme:          cattleURL.Scheme,
		httpListenAddr:  httpListenAddr,
		httpsListenAddr: httpsListenAddr,
		certFile:        certFile,
		keyFile:         keyFile,
	}, nil
}

func (p *Proxy) ListenAndServe() error {
	group, _ := errgroup.WithContext(context.Background())

	if p.httpListenAddr == "" {
		logrus.Infof("Not proxying k8s http")
	} else {
		logrus.Infof("Listening on %s for k8s proxy", p.httpListenAddr)
		group.Go(func() error {
			return http.ListenAndServe(p.httpListenAddr, p.Handler())
		})

	}

	if p.httpsListenAddr == "" || p.keyFile == "" || p.certFile == "" {
		logrus.Infof("Not proxying k8s https")
	} else {
		group.Go(func() error {
			logrus.Infof("Listening on %s for tls k8s proxy", p.httpsListenAddr)
			return http.ListenAndServeTLS(p.httpsListenAddr, p.certFile, p.keyFile, p.Handler())
		})
	}

	return group.Wait()
}

func (p *Proxy) Handler() http.Handler {
	return http.HandlerFunc(p.Handle)
}

func (p *Proxy) Handle(rw http.ResponseWriter, req *http.Request) {
	unmodified := req.URL.Path
	req.URL.Path = p.prefix + req.URL.Path
	req.Host = p.cattleAddr

	logrus.Debugf("Proxying %s %s %s => %s", req.Method, req.Host, unmodified, req.URL.Path)
	if websocket.ShouldProxy(req) {
		websocket.Proxy(p.scheme, p.cattleAddr, rw, req)
	} else {
		p.reverseProxy.ServeHTTP(rw, req)
	}
}
