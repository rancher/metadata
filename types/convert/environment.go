package convert

import (
	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/metadata/content"
	"github.com/rancher/metadata/types"
)

type EnvironmentWrapper struct {
	Client      content.Client
	Environment *client.EnvironmentInfo
	Store       content.Store
}

func NewEnvironmentObject(obj interface{}, c content.Client, store content.Store) content.Object {
	return &WrappedObject{
		Wrapped: &EnvironmentWrapper{
			Client:      c,
			Environment: obj.(*client.EnvironmentInfo),
			Store:       store,
		},
	}
}

func (c *EnvironmentWrapper) wrapped() interface{} {
	result := &types.EnvironmentResponse{
		Name:       c.Environment.Name,
		ExternalID: c.Environment.ExternalId,
		System:     c.Environment.System,
		UUID:       c.Environment.Uuid,

		Containers: c.Store.ByEnvironment(content.ContainerType, c.Client, c.Environment.Uuid),
		Services:   c.Store.ByEnvironment(content.ServiceType, c.Client, c.Environment.Uuid),
		Networks:   c.Store.ByEnvironment(content.NetworkType, c.Client, c.Environment.Uuid),
		Hosts:      c.Store.ByEnvironment(content.HostType, c.Client, c.Environment.Uuid),
		Stacks:     c.Store.ByEnvironment(content.StackType, c.Client, c.Environment.Uuid),
		Version:    c.Store.Version(),
	}
	return result
}
