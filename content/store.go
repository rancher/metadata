package content

import (
	"github.com/rancher/go-rancher/v3"
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
	ObjectFactories = map[ObjectType]ObjectFactory{}
)

type ObjectType string

type ObjectFactory func(obj interface{}, client Client, store Store) Object

type IDResolution interface {
	IDtoUUID(objectType ObjectType, id string) string
}

type Store interface {
	IDResolution

	// Return types are all Object because the generic object walker for the API
	// does not deal with concrete types such as []ContainerWrapper, EnvironmentWrapper
	Environment(client Client) Object

	ByEnvironment(objectType ObjectType, client Client, environmentUUID string) []Object
	ByStack(objectType ObjectType, client Client, stackUUID string) []Object

	Object(uuid string, client Client) Object

	ServiceByID(id string) *client.ServiceInfo
	StackByID(id string) *client.StackInfo
	NetworkByID(id string) *client.NetworkInfo
	HostByID(id string) *client.HostInfo
	ContainerByID(id string) *client.InstanceInfo
	EnvironmentByUUID(environmentUUID string) *client.EnvironmentInfo

	ServiceByName(environmentUUID, stackName, name string) *client.ServiceInfo
	ContainerByName(environmentUUID, stackName, name string) *client.InstanceInfo

	// Self
	SelfContainer(client Client) *client.InstanceInfo
	SelfHost(client Client) Object

	Reload(all map[string]interface{})

	Add(val map[string]interface{})
	Remove(val map[string]interface{})

	Version() string

	WaitChanged()
}
