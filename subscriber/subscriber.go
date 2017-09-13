package subscriber

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"
	"github.com/rancher/event-subscriber/events"
	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/metadata/content"
)

const data = "./metadata-content"
const generationFile = "generation"

type Subscriber struct {
	sync.Mutex

	client     *client.RancherClient
	opts       *client.ClientOpts
	store      content.Store
	router     *events.EventRouter
	generation string
}

func NewSubscriber(opts *client.ClientOpts, store content.Store) (*Subscriber, error) {
	s := &Subscriber{
		opts:  opts,
		store: store,
	}

	if err := s.restore(); err != nil {
		s.clearGeneration()
		return nil, err
	}

	return s, nil
}

func (s *Subscriber) Start() {
	for {
		client, err := client.NewRancherClient(s.opts)
		if err == nil {
			s.client = client
			break
		}

		logrus.Errorf("Error creating client: %v", err)
		time.Sleep(5 * time.Second)
	}

	for {
		router, err := events.NewEventRouter(s.client, 2, map[string]events.EventHandler{
			"metadata.sync": events.EventHandler(s.sync),
		})
		if err == nil {
			s.router = router
			break
		}
		logrus.Errorf("Error creating event router: %v", err)
		time.Sleep(5 * time.Second)
	}

	for {
		err := s.router.Start(nil)
		if err != nil {
			logrus.Errorf("Error subscribing: %v", err)
		}
		time.Sleep(2 * time.Second)
	}
}

func (s *Subscriber) sync(event *events.Event, c *client.RancherClient) error {
	s.Lock()
	defer s.Unlock()

	start := time.Now()

	reload := false
	request := &client.MetadataSyncRequest{}
	if err := mapstructure.Decode(event.Data["metadataSyncRequest"], request); err != nil {
		return err
	}
	if request.Full {
		s.store.Reload(request.Updates)
		s.generation = request.Generation
		if err := s.save(request); err != nil {
			return err
		}
	} else if s.generation == request.Generation {
		for _, obj := range request.Updates {
			s.store.Add(obj.(map[string]interface{}))
		}

		for _, obj := range request.Removes {
			s.store.Remove(obj.(map[string]interface{}))
		}
		if err := s.save(request); err != nil {
			return err
		}
	} else {
		s.Reload()
		reload = true
	}

	publishStart := time.Now()
	_, err := c.Publish.Create(&client.Publish{
		Name:       event.ReplyTo,
		PreviousId: event.ID,
		Data: map[string]interface{}{
			"reload": reload,
		},
	})

	end := time.Now()
	logrus.Debugf("Processing done after %v and %v to post", end.Sub(start), publishStart.Sub(publishStart))

	return err
}

func (s *Subscriber) Reload() {
	logrus.Info("Requesting reload, on next event")
	s.generation = ""
}

func (s *Subscriber) save(request *client.MetadataSyncRequest) error {
	base := path.Join(data, request.Generation)
	if request.Full {
		if err := os.RemoveAll(base); !os.IsNotExist(err) && err != nil {
			return err
		}
	}

	var lastError error
	for uuid, obj := range request.Updates {
		if err := s.writeAtomic(base, uuid, obj); err != nil {
			lastError = err
			continue
		}
	}

	for uuid := range request.Removes {
		if err := os.Remove(path.Join(base, uuid)); !os.IsNotExist(err) && err != nil {
			lastError = err
			continue
		}
	}

	if request.Full {
		if err := s.writeAtomic(data, generationFile, request.Generation); err != nil {
			return err
		}

		logrus.Infof("New generation %s", request.Generation)

		files, err := ioutil.ReadDir(data)
		if err != nil {
			return err
		}

		for _, file := range files {
			if file.Name() == generationFile || file.Name() == request.Generation {
				continue
			}

			fullName := path.Join(data, file.Name())
			logrus.Infof("Deleting %s", fullName)
			os.RemoveAll(fullName)
		}
	}

	return lastError
}

func (s *Subscriber) restore() error {
	generation, err := ioutil.ReadFile(path.Join(data, generationFile))
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	logrus.Debugf("Restoring generation %s", generation)

	base := path.Join(data, string(generation))
	files, err := ioutil.ReadDir(base)
	if err != nil {
		return err
	}

	vals := map[string]interface{}{}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".tmp") {
			continue
		}

		obj := map[string]interface{}{}
		filename := path.Join(base, file.Name())
		logrus.Debugf("Loading %s", filename)
		f, err := os.Open(filename)
		if err != nil {
			return err
		}

		err = json.NewDecoder(f).Decode(&obj)
		f.Close()
		if err != nil {
			return err
		}

		uuid, _ := obj["uuid"].(string)
		if uuid != "" {
			vals[uuid] = obj
		}
	}

	s.store.Reload(vals)
	s.generation = string(generation)
	logrus.Debugf("Generation %s", s.generation)

	return nil
}

func (s *Subscriber) clearGeneration() error {
	err := os.Remove(path.Join(data, generationFile))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (s *Subscriber) writeAtomic(base, uuid string, obj interface{}) error {
	file := path.Join(base, uuid)
	temp := file + ".tmp"

	logrus.Debugf("Writing %s", file)

	os.MkdirAll(path.Dir(temp), 0700)
	f, err := os.Create(temp)
	if err != nil {
		return err
	}

	if s, ok := obj.(string); ok {
		if _, err := f.Write([]byte(s)); err != nil {
			f.Close()
			os.Remove(temp)
			return err
		}
	} else if err := json.NewEncoder(f).Encode(obj); err != nil {
		f.Close()
		os.Remove(temp)
		return err
	}

	if err := f.Close(); err != nil {
		os.Remove(temp)
		return err
	}

	return os.Rename(temp, file)
}
