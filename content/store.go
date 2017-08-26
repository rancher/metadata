package content

import (
	"github.com/rancher/metadata/types"
)

const (
	ContainerType   = ObjectType("instance")
	ServiceType     = ObjectType("service")
	StackType       = ObjectType("stack")
	NetworkType     = ObjectType("network")
	HostType        = ObjectType("host")
	EnvironmentType = ObjectType("environment")
)

var (
	Types = []ObjectType{
		ContainerType,
		ServiceType,
		StackType,
		NetworkType,
		HostType,
		EnvironmentType,
	}
	ObjectFactories = map[ObjectType]ObjectFactory{
		ContainerType:   NewContainerObject,
		ServiceType:     NewServiceObject,
		StackType:       NewStackObject,
		NetworkType:     NewNetworkObject,
		HostType:        NewHostObject,
		EnvironmentType: NewEnvironmentObject,
	}
)

type ObjectType string

type ObjectFactory func(obj interface{}, client Client, store Store) types.Object

type IDResolution interface {
	IDtoUUID(objectType ObjectType, id string) string
}

type ServiceIndexed interface {
	GetServiceID() string
}

type EnvironmentIndexed interface {
	GetEnvironmentUUID() string
}

type Store interface {
	IDResolution

	// Return types are all Object because the generic object walker for the API
	// does not deal with concrete types such as []Container, Environment
	Environment(client Client) types.Object

	ByService(objectType ObjectType, client Client, serviceUUID string) []types.Object
	ByEnvironment(objectType ObjectType, client Client, environmentUUID string) []types.Object

	Reload(all map[string]interface{})

	Add(val map[string]interface{})
	Remove(val map[string]interface{})

	Version() string

	WaitChanged()
}
