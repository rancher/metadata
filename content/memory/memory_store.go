package memory

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"
	"github.com/rancher/metadata/content"
	"github.com/rancher/metadata/types"
	"golang.org/x/sync/syncmap"
)

//TODO: replace this stupidity with RDBMS or an indexed kv

type storeData struct {
	objects syncmap.Map // string(objectType) => string(uuid) => *types.(EnvironmentWrapper|Instance|...)
	idMap   syncmap.Map // string(type:id) => string(uuid)
	all     syncmap.Map // string(uuid) => *types.(EnvironmentWrapper|Instance|...)
}

type Store struct {
	version int
	d       *storeData
	cond    *sync.Cond
}

type objectSliceWrapper struct {
	slice []types.Object
}

func NewMemoryStore(ctx context.Context) *Store {
	m := &Store{
		d:       newStoreData(),
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

func newStoreData() *storeData {
	data := &storeData{}
	for _, objectType := range content.Types {
		data.objects.Store(objectType, &syncmap.Map{})
	}
	return data
}

func (m *Store) getObjectMap(objectType content.ObjectType) *syncmap.Map {
	val, _ := m.d.objects.Load(objectType)
	return val.(*syncmap.Map)
}

func (m *Store) Environment(c content.Client) types.Object {
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

func (m *Store) ServiceByName(environmentUUID, stackName, name string) *types.Service {
	var result *types.Service
	var stack *types.Stack

	m.getObjectMap(content.StackType).Range(func(key, value interface{}) bool {
		obj := value.(*types.Stack)
		if obj.EnvironmentUUID == environmentUUID && strings.EqualFold(obj.Name, stackName) {
			stack = obj
			return false
		}
		return true
	})

	if stack == nil {
		return nil
	}

	m.getObjectMap(content.ServiceType).Range(func(key, value interface{}) bool {
		obj := value.(*types.Service)
		if obj.EnvironmentUUID == environmentUUID && obj.StackID == stack.ID && strings.EqualFold(obj.Name, name) {
			result = obj
			return false
		}
		return true
	})

	return result

}

func (m *Store) ContainerByName(environmentUUID, stackName, name string) *types.Container {
	var result *types.Container
	var stack *types.Stack

	m.getObjectMap(content.StackType).Range(func(key, value interface{}) bool {
		obj := value.(*types.Stack)
		if obj.EnvironmentUUID == environmentUUID && strings.EqualFold(obj.Name, stackName) {
			stack = obj
			return false
		}
		return true
	})

	if stack == nil {
		return nil
	}

	m.getObjectMap(content.ContainerType).Range(func(key, value interface{}) bool {
		obj := value.(*types.Container)
		if obj.EnvironmentUUID == environmentUUID && obj.StackID == stack.ID && strings.EqualFold(obj.Name, name) {
			result = obj
			return false
		}
		return true
	})

	return result
}

func (m *Store) ByStack(objectType content.ObjectType, c content.Client, stackUUID string) []types.Object {
	result := objectSliceWrapper{}
	objects := m.getObjectMap(objectType)

	objects.Range(func(key, value interface{}) bool {
		indexed, ok := value.(content.StackIndexed)
		if ok {
			if m.IDtoUUID(content.StackType, indexed.GetStackID()) == stackUUID {
				result.slice = append(result.slice, m.newObject(objectType, value, c))
			}
		}
		return true
	})

	return result.slice
}

func (m *Store) newObject(objectType content.ObjectType, obj interface{}, c content.Client) types.Object {
	return content.ObjectFactories[objectType](obj, c, m)
}

func (m *Store) ByEnvironment(objectType content.ObjectType, c content.Client, environmentUUID string) []types.Object {
	result := objectSliceWrapper{}

	env, ok := m.getEnv(environmentUUID)
	if !ok {
		return result.slice
	}

	m.getObjectMap(objectType).Range(func(key, value interface{}) bool {
		indexed, ok := value.(content.EnvironmentIndexed)
		if ok {
			if env.System || indexed.GetEnvironmentUUID() == environmentUUID {
				result.slice = append(result.slice, m.newObject(objectType, value, c))
			}
		}
		return true
	})

	return result.slice
}

func (m *Store) getEnv(uuid string) (*types.Environment, bool) {
	val, _ := m.d.all.Load(uuid)
	env, ok := val.(*types.Environment)
	return env, ok
}

func (m *Store) instanceByIP(clientIP string) *types.Container {
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

func (m *Store) putIDtoUUID(objectType content.ObjectType, id, uuid string) {
	m.d.idMap.Store(fmt.Sprintf("%s:%s", objectType, id), uuid)
}

func (m *Store) removeIDtoUUID(objectType content.ObjectType, id string) {
	m.d.idMap.Delete(fmt.Sprintf("%s:%s", objectType, id))
}

func (m *Store) IDtoUUID(objectType content.ObjectType, id string) string {
	uuid, _ := m.d.idMap.Load(fmt.Sprintf("%s:%s", objectType, id))
	s, _ := uuid.(string)
	return s
}

func (m *Store) WaitChanged() {
	m.cond.L.Lock()
	m.cond.Wait()
	m.cond.L.Unlock()
}

func (m *Store) Changed() {
	m.cond.Broadcast()
}

func (m *Store) Add(val map[string]interface{}) {
	infoType, _ := val["infoType"].(string)
	id, _ := val["infoTypeId"].(string)
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

func (m *Store) Remove(val map[string]interface{}) {
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

func (m *Store) SelfContainer(c content.Client) *types.Container {
	return m.instanceByIP(c.IP)
}

func (m *Store) SelfHost(c content.Client) types.Object {
	hostname, err := os.Hostname()
	if err != nil {
		logrus.Errorf("Failed to get hostname: %v", err)
		return nil
	}

	i := strings.Index(hostname, ".")
	if i > 0 {
		hostname = hostname[i:]
	}

	var result *types.Container

	m.getObjectMap(content.ContainerType).Range(func(key, value interface{}) bool {
		instance := value.(*types.Container)
		if strings.HasPrefix(instance.ExternalID, hostname) {
			result = instance
			return false
		}
		return true
	})

	if result == nil {
		return nil
	}

	return m.Object(m.IDtoUUID(content.HostType, result.HostID), c)
}

func (m *Store) Object(uuid string, c content.Client) types.Object {
	var result types.Object
	m.d.objects.Range(func(key, value interface{}) bool {
		data := value.(*syncmap.Map)
		found, ok := data.Load(uuid)
		if ok {
			result = m.newObject(key.(content.ObjectType), found, c)
			return false
		}
		return true
	})

	return result
}

func (m *Store) Reload(vals map[string]interface{}) {
	o := NewMemoryStore(nil)

	for _, rawVal := range vals {
		o.Add(rawVal.(map[string]interface{}))
	}

	m.d = o.d
	m.version = o.version
}

func (m *Store) Version() string {
	return strconv.Itoa(m.version)
}

func (m *Store) ServiceByID(id string) *types.Service {
	val, ok := m.getObjectMap(content.ServiceType).Load(m.IDtoUUID(content.ServiceType, id))
	if ok {
		return val.(*types.Service)
	}
	return nil
}

func (m *Store) StackByID(id string) *types.Stack {
	val, ok := m.getObjectMap(content.StackType).Load(m.IDtoUUID(content.StackType, id))
	if ok {
		return val.(*types.Stack)
	}
	return nil
}

func (m *Store) HostByID(id string) *types.Host {
	val, ok := m.getObjectMap(content.HostType).Load(m.IDtoUUID(content.HostType, id))
	if ok {
		return val.(*types.Host)
	}
	return nil
}

func (m *Store) NetworkByID(id string) *types.Network {
	val, ok := m.getObjectMap(content.NetworkType).Load(m.IDtoUUID(content.NetworkType, id))
	if ok {
		return val.(*types.Network)
	}
	return nil
}

func (m *Store) ContainerByID(id string) *types.Container {
	val, ok := m.getObjectMap(content.ContainerType).Load(m.IDtoUUID(content.ContainerType, id))
	if ok {
		return val.(*types.Container)
	}
	return nil
}

func (m *Store) EnvironmentByUUID(uuid string) *types.Environment {
	val, ok := m.getObjectMap(content.EnvironmentType).Load(uuid)
	if ok {
		return val.(*types.Environment)
	}
	return nil
}
