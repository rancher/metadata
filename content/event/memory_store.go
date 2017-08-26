package event

import (
	"context"
	"fmt"
	"sync"
	"time"

	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"
	"github.com/rancher/metadata/content"
	"github.com/rancher/metadata/types"
	"golang.org/x/sync/syncmap"
)

type memoryStoreData struct {
	objects syncmap.Map // string(objectType) => string(uuid) => *types.(Environment|Instance|...)
	idMap   syncmap.Map // string(type:id) => string(uuid)
	all     syncmap.Map // string(uuid) => *types.(Environment|Instance|...)
}

type MemoryStore struct {
	version int
	d       *memoryStoreData
	cond    *sync.Cond
}

type objectSliceWrapper struct {
	slice []types.Object
}

func NewMemoryStore(ctx context.Context) *MemoryStore {
	m := &MemoryStore{
		d:       newMemoryStoreData(),
		cond:    sync.NewCond(&sync.Mutex{}),
		version: time.Now().Nanosecond(),
	}

	if ctx != nil {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
					m.Changed()
				}
			}
		}()
	}

	return m
}

func newMemoryStoreData() *memoryStoreData {
	data := &memoryStoreData{}
	for _, objectType := range content.Types {
		data.objects.Store(objectType, &syncmap.Map{})
	}
	return data
}

func (m *MemoryStore) getObjectMap(objectType content.ObjectType) *syncmap.Map {
	val, _ := m.d.objects.Load(objectType)
	return val.(*syncmap.Map)
}

func (m *MemoryStore) Environment(c content.Client) types.Object {
	var result *types.Environment

	environments := m.getObjectMap(content.EnvironmentType)
	container := m.instanceByIP(c.IP)
	if container == nil {
		environments.Range(func(key, value interface{}) bool {
			if value.(*types.Environment).System {
				result = value.(*types.Environment)
				return false
			}
			return true
		})
	} else {
		environments.Range(func(key, value interface{}) bool {
			if value.(*types.Environment).UUID == container.EnvironmentUUID {
				result = value.(*types.Environment)
				return false
			}
			return true
		})
	}

	if result == nil {
		return nil
	}

	return content.NewEnvironmentObject(result, c, m)
}

func (m *MemoryStore) ByService(objectType content.ObjectType, c content.Client, serviceUUID string) []types.Object {
	result := objectSliceWrapper{}
	objects := m.getObjectMap(objectType)

	objects.Range(func(key, value interface{}) bool {
		indexed, ok := value.(content.ServiceIndexed)
		if ok {
			if m.IDtoUUID(content.ServiceType, indexed.GetServiceID()) == serviceUUID {
				result.slice = append(result.slice, m.newObject(objectType, value, c))
			}
		}
		return true
	})

	return result.slice
}

func (m *MemoryStore) newObject(objectType content.ObjectType, obj interface{}, c content.Client) types.Object {
	return content.ObjectFactories[objectType](obj, c, m)
}

func (m *MemoryStore) ByEnvironment(objectType content.ObjectType, c content.Client, environmentUUID string) []types.Object {
	result := objectSliceWrapper{}

	env, ok := m.getEnv(environmentUUID)
	if !ok {
		return result.slice
	}

	m.getObjectMap(objectType).Range(func(key, value interface{}) bool {
		indexed, ok := value.(content.EnvironmentIndexed)
		if ok {
			if env.System || indexed.GetEnvironmentUUID() == environmentUUID {
				fmt.Println(m.newObject(objectType, value, c))
				result.slice = append(result.slice, m.newObject(objectType, value, c))
			}
		}
		return true
	})

	return result.slice
}

func (m *MemoryStore) getEnv(uuid string) (*types.Environment, bool) {
	val, _ := m.d.all.Load(uuid)
	env, ok := val.(*types.Environment)
	return env, ok
}

func (m *MemoryStore) instanceByIP(clientIP string) *types.Container {
	var result *types.Container

	m.getObjectMap(content.ContainerType).Range(func(key, value interface{}) bool {
		instance := value.(*types.Container)
		if instance.PrimaryIP == clientIP {
			result = instance
			return false
		}
		return true
	})

	return result
}

func (m *MemoryStore) putIDtoUUID(objectType content.ObjectType, id, uuid string) {
	m.d.idMap.Store(fmt.Sprintf("%s:%s", objectType, id), uuid)
}

func (m *MemoryStore) removeIDtoUUID(objectType content.ObjectType, id string) {
	m.d.idMap.Delete(fmt.Sprintf("%s:%s", objectType, id))
}

func (m *MemoryStore) IDtoUUID(objectType content.ObjectType, id string) string {
	uuid, _ := m.d.idMap.Load(fmt.Sprintf("%s:%s", objectType, id))
	s, _ := uuid.(string)
	return s
}

func (m *MemoryStore) WaitChanged() {
	m.cond.L.Lock()
	m.cond.Wait()
	m.cond.L.Unlock()
}

func (m *MemoryStore) Changed() {
	m.cond.Broadcast()
}

func (m *MemoryStore) Add(val map[string]interface{}) {
	infoType, _ := val["infoType"].(string)
	id, _ := val["id"].(string)
	uuid, _ := val["uuid"].(string)

	logrus.Infof("Adding %s %s:%s", uuid, infoType, id)

	if id == "" || id == "0" || uuid == "" {
		return
	}

	var rawVal interface{}
	objectType := content.ObjectType(infoType)

	switch objectType {
	case content.ContainerType:
		rawVal = &types.Container{}
	case content.ServiceType:
		rawVal = &types.Service{}
	case content.StackType:
		rawVal = &types.Stack{}
	case content.NetworkType:
		rawVal = &types.Network{}
	case content.HostType:
		rawVal = &types.Host{}
	case content.EnvironmentType:
		rawVal = &types.Environment{}
	}

	if rawVal != nil {
		obj := decodeAndLog(val, rawVal)
		m.getObjectMap(objectType).Store(uuid, obj)
		m.putIDtoUUID(objectType, id, uuid)

		m.d.all.Store(uuid, obj)
		m.version++
		m.Changed()
	}
}

func decodeAndLog(m interface{}, rawVal interface{}) interface{} {
	err := mapstructure.Decode(m, rawVal)
	if err != nil {
		logrus.Errorf("Failed to unmarshal: %v %v", err, m)
	}
	return rawVal
}

func (m *MemoryStore) Remove(val map[string]interface{}) {
	id := fmt.Sprint(val["id"])
	uuid, _ := val["uuid"].(string)
	infoType, _ := val["infoType"].(string)

	logrus.Infof("Removing %s %s:%s", uuid, infoType, id)
	if uuid == "" || id == "" || infoType == "" {
		return
	}

	objectType := content.ObjectType(infoType)

	objectMap := m.getObjectMap(objectType)
	if objectMap == nil {
		return
	}

	objectMap.Delete(uuid)
	m.removeIDtoUUID(objectType, id)
	m.d.all.Delete(uuid)
	m.version++
	m.Changed()
}

func (m *MemoryStore) Reload(vals map[string]interface{}) {
	o := NewMemoryStore(nil)

	for _, rawVal := range vals {
		o.Add(rawVal.(map[string]interface{}))
	}

	m.d = o.d
	m.version = o.version
}

func (m *MemoryStore) Version() string {
	return strconv.Itoa(m.version)
}
