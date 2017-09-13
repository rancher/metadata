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
	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/metadata/content"
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
	slice []content.Object
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

func (m *Store) Environment(c content.Client) content.Object {
	var result *client.EnvironmentInfo

	environments := m.getObjectMap(content.EnvironmentType)
	container := m.instanceByIP(c.IP)
	if container == nil {
		environments.Range(func(key, value interface{}) bool {
			if value.(*client.EnvironmentInfo).System {
				result = value.(*client.EnvironmentInfo)
				return false
			}
			return true
		})
	} else {
		environments.Range(func(key, value interface{}) bool {
			if value.(*client.EnvironmentInfo).Uuid == container.EnvironmentUuid {
				result = value.(*client.EnvironmentInfo)
				return false
			}
			return true
		})
	}

	if result == nil {
		return nil
	}

	return content.ObjectFactories[content.EnvironmentType](result, c, m)
}

func (m *Store) ServiceByName(environmentUUID, stackName, name string) *client.ServiceInfo {
	var result *client.ServiceInfo
	var stack *client.StackInfo

	m.getObjectMap(content.StackType).Range(func(key, value interface{}) bool {
		obj := value.(*client.StackInfo)
		if obj.EnvironmentUuid == environmentUUID && strings.EqualFold(obj.Name, stackName) {
			stack = obj
			return false
		}
		return true
	})

	if stack == nil {
		return nil
	}

	m.getObjectMap(content.ServiceType).Range(func(key, value interface{}) bool {
		obj := value.(*client.ServiceInfo)
		if obj.EnvironmentUuid == environmentUUID && obj.StackId == stack.InfoTypeId && strings.EqualFold(obj.Name, name) {
			result = obj
			return false
		}
		return true
	})

	return result

}

func (m *Store) ContainerByName(environmentUUID, stackName, name string) *client.InstanceInfo {
	var result *client.InstanceInfo
	var stack *client.StackInfo

	m.getObjectMap(content.StackType).Range(func(key, value interface{}) bool {
		obj := value.(*client.StackInfo)
		if obj.EnvironmentUuid == environmentUUID && strings.EqualFold(obj.Name, stackName) {
			stack = obj
			return false
		}
		return true
	})

	if stack == nil {
		return nil
	}

	m.getObjectMap(content.ContainerType).Range(func(key, value interface{}) bool {
		obj := value.(*client.InstanceInfo)
		if obj.EnvironmentUuid == environmentUUID && obj.StackId == stack.Id && strings.EqualFold(obj.Name, name) {
			result = obj
			return false
		}
		return true
	})

	return result
}

func (m *Store) ByStack(objectType content.ObjectType, c content.Client, stackUUID string) []content.Object {
	result := objectSliceWrapper{}
	objects := m.getObjectMap(objectType)

	objects.Range(func(key, value interface{}) bool {
		stackID, ok := getString(value, "StackId")
		if ok {
			if m.IDtoUUID(content.StackType, stackID) == stackUUID {
				result.slice = append(result.slice, m.newObject(objectType, value, c))
			}
		}
		return true
	})

	return result.slice
}

func getString(obj interface{}, key string) (string, bool) {
	val, ok := content.GetValue(obj, key)
	if !ok {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

func (m *Store) newObject(objectType content.ObjectType, obj interface{}, c content.Client) content.Object {
	return content.ObjectFactories[objectType](obj, c, m)
}

func (m *Store) ByEnvironment(objectType content.ObjectType, c content.Client, environmentUUID string) []content.Object {
	result := objectSliceWrapper{}

	env, ok := m.getEnv(environmentUUID)
	if !ok {
		return result.slice
	}

	m.getObjectMap(objectType).Range(func(key, value interface{}) bool {
		testUUID, ok := getString(value, "EnvironmentUuid")
		if ok {
			if env.System || testUUID == environmentUUID {
				result.slice = append(result.slice, m.newObject(objectType, value, c))
			}
		}
		return true
	})

	return result.slice
}

func (m *Store) getEnv(uuid string) (*client.EnvironmentInfo, bool) {
	val, _ := m.d.all.Load(uuid)
	env, ok := val.(*client.EnvironmentInfo)
	return env, ok
}

func (m *Store) instanceByIP(clientIP string) *client.InstanceInfo {
	var result *client.InstanceInfo

	m.getObjectMap(content.ContainerType).Range(func(key, value interface{}) bool {
		instance := value.(*client.InstanceInfo)
		if instance.PrimaryIp == clientIP {
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

	// copy map since we are changing a value
	val = copyMap(val)
	val["id"] = id

	var rawVal interface{}
	objectType := content.ObjectType(infoType)

	switch objectType {
	case content.ContainerType:
		rawVal = &client.InstanceInfo{}
	case content.ServiceType:
		rawVal = &client.ServiceInfo{}
	case content.StackType:
		rawVal = &client.StackInfo{}
	case content.NetworkType:
		rawVal = &client.NetworkInfo{}
	case content.HostType:
		rawVal = &client.HostInfo{}
	case content.EnvironmentType:
		rawVal = &client.EnvironmentInfo{}
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

func copyMap(val map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for k, v := range val {
		result[k] = v
	}
	return result
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

func (m *Store) SelfContainer(c content.Client) *client.InstanceInfo {
	return m.instanceByIP(c.IP)
}

func (m *Store) SelfHost(c content.Client) content.Object {
	hostname, err := os.Hostname()
	if err != nil {
		logrus.Errorf("Failed to get hostname: %v", err)
		return nil
	}

	i := strings.Index(hostname, ".")
	if i > 0 {
		hostname = hostname[i:]
	}

	var result *client.InstanceInfo

	m.getObjectMap(content.ContainerType).Range(func(key, value interface{}) bool {
		instance := value.(*client.InstanceInfo)
		if strings.HasPrefix(instance.ExternalId, hostname) {
			result = instance
			return false
		}
		return true
	})

	if result == nil {
		return nil
	}

	return m.Object(m.IDtoUUID(content.HostType, result.HostId), c)
}

func (m *Store) Object(uuid string, c content.Client) content.Object {
	var result content.Object
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

func (m *Store) ServiceByID(id string) *client.ServiceInfo {
	val, ok := m.getObjectMap(content.ServiceType).Load(m.IDtoUUID(content.ServiceType, id))
	if ok {
		return val.(*client.ServiceInfo)
	}
	return nil
}

func (m *Store) StackByID(id string) *client.StackInfo {
	val, ok := m.getObjectMap(content.StackType).Load(m.IDtoUUID(content.StackType, id))
	if ok {
		return val.(*client.StackInfo)
	}
	return nil
}

func (m *Store) HostByID(id string) *client.HostInfo {
	val, ok := m.getObjectMap(content.HostType).Load(m.IDtoUUID(content.HostType, id))
	if ok {
		return val.(*client.HostInfo)
	}
	return nil
}

func (m *Store) NetworkByID(id string) *client.NetworkInfo {
	val, ok := m.getObjectMap(content.NetworkType).Load(m.IDtoUUID(content.NetworkType, id))
	if ok {
		return val.(*client.NetworkInfo)
	}
	return nil
}

func (m *Store) ContainerByID(id string) *client.InstanceInfo {
	val, ok := m.getObjectMap(content.ContainerType).Load(m.IDtoUUID(content.ContainerType, id))
	if ok {
		return val.(*client.InstanceInfo)
	}
	return nil
}

func (m *Store) EnvironmentByUUID(uuid string) *client.EnvironmentInfo {
	val, ok := m.getObjectMap(content.EnvironmentType).Load(uuid)
	if ok {
		return val.(*client.EnvironmentInfo)
	}
	return nil
}
