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

type StackIndexed interface {
	GetStackID() string
}

type EnvironmentIndexed interface {
	GetEnvironmentUUID() string
}

type Store interface {
	IDResolution

	// Return types are all Object because the generic object walker for the API
	// does not deal with concrete types such as []ContainerWrapper, EnvironmentWrapper
	Environment(client Client) types.Object

	ByEnvironment(objectType ObjectType, client Client, environmentUUID string) []types.Object
	ByStack(objectType ObjectType, client Client, stackUUID string) []types.Object

	Object(uuid string, client Client) types.Object

	ServiceByID(id string) *types.Service
	StackByID(id string) *types.Stack
	NetworkByID(id string) *types.Network
	HostByID(id string) *types.Host
	ContainerByID(id string) *types.Container
	EnvironmentByUUID(environmentUUID string) *types.Environment

	ServiceByName(environmentUUID, stackName, name string) *types.Service
	ContainerByName(environmentUUID, stackName, name string) *types.Container

	// Self
	SelfContainer(client Client) *types.Container
	SelfHost(client Client) types.Object

	Reload(all map[string]interface{})

	Add(val map[string]interface{})
	Remove(val map[string]interface{})

	Version() string

	WaitChanged()
}
