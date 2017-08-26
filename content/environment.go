package content

import (
	"github.com/rancher/metadata/types"
)

type Environment struct {
	Client      Client
	Environment *types.Environment
	Store       Store
}

func NewEnvironmentObject(obj interface{}, client Client, store Store) types.Object {
	return &WrappedObject{
		Wrapped: &Environment{
			Client:      client,
			Environment: obj.(*types.Environment),
			Store:       store,
		},
	}
}

func (c *Environment) wrapped() interface{} {
	//switch c.Client.Version {
	//case V3:
	result := *c.Environment
	result.EnvironmentChildren = c.children()
	result.Version = c.Store.Version()
	return result
	//}
	//
	//return nil
}

func (c *Environment) children() types.EnvironmentChildren {
	return types.EnvironmentChildren{
		Containers: c.Store.ByEnvironment(ContainerType, c.Client, c.Environment.UUID),
		Services:   c.Store.ByEnvironment(ServiceType, c.Client, c.Environment.UUID),
		Networks:   c.Store.ByEnvironment(NetworkType, c.Client, c.Environment.UUID),
		Hosts:      c.Store.ByEnvironment(HostType, c.Client, c.Environment.UUID),
		Stacks:     c.Store.ByEnvironment(StackType, c.Client, c.Environment.UUID),
	}
}
