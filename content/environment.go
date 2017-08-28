package content

import (
	"github.com/rancher/metadata/types"
)

type EnvironmentWrapper struct {
	Client      Client
	Environment *types.Environment
	Store       Store
}

func NewEnvironmentObject(obj interface{}, client Client, store Store) types.Object {
	return &WrappedObject{
		Wrapped: &EnvironmentWrapper{
			Client:      client,
			Environment: obj.(*types.Environment),
			Store:       store,
		},
	}
}

func (c *EnvironmentWrapper) wrapped() interface{} {
	result := &types.EnvironmentResponse{
		Environment: *c.Environment,
		EnvironmentDynamic: types.EnvironmentDynamic{
			Containers: c.Store.ByEnvironment(ContainerType, c.Client, c.Environment.UUID),
			Services:   c.Store.ByEnvironment(ServiceType, c.Client, c.Environment.UUID),
			Networks:   c.Store.ByEnvironment(NetworkType, c.Client, c.Environment.UUID),
			Hosts:      c.Store.ByEnvironment(HostType, c.Client, c.Environment.UUID),
			Stacks:     c.Store.ByEnvironment(StackType, c.Client, c.Environment.UUID),
			Version:    c.Store.Version(),
		},
	}
	return result
}
