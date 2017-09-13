package convert

import (
	"github.com/rancher/metadata/content"
	"github.com/rancher/metadata/types"
)

type Self struct {
	Client content.Client
	Store  content.Store
}

func NewSelfObject(version, ip string, store content.Store) content.Object {
	return &WrappedObject{
		Wrapped: &Self{
			Client: content.Client{
				Version: version,
				IP:      ip,
			},
			Store: store,
		},
	}
}

func (c *Self) wrapped() interface{} {
	container := c.Store.SelfContainer(c.Client)
	if container == nil {
		return types.MetadataSelf{
			Host: c.Store.SelfHost(c.Client),
		}
	}

	s := c.Store
	o := c.Store.Object
	self := types.MetadataSelf{
		Container:   o(container.Uuid, c.Client),
		Service:     o(s.IDtoUUID(content.ServiceType, container.ServiceId), c.Client),
		Host:        o(s.IDtoUUID(content.HostType, container.HostId), c.Client),
		Environment: o(container.EnvironmentUuid, c.Client),
		Network:     o(s.IDtoUUID(content.ContainerType, container.NetworkId), c.Client),
		Stack:       o(s.IDtoUUID(content.StackType, container.StackId), c.Client),
	}

	wrapped, ok := self.Service.(*WrappedObject)
	if ok {
		serviceWrapper, ok := wrapped.Wrapped.(*ServiceWrapper)
		if ok {
			serviceWrapper.IncludeToken = true
		}
	}

	return self
}
