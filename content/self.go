package content

import (
	"github.com/rancher/metadata/types"
)

type Self struct {
	Client Client
	Store  Store
}

func NewSelfObject(version, ip string, store Store) types.Object {
	return &WrappedObject{
		Wrapped: &Self{
			Client: Client{
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
		Container:   o(container.UUID, c.Client),
		Service:     o(s.IDtoUUID(ServiceType, container.ServiceID), c.Client),
		Host:        o(s.IDtoUUID(HostType, container.HostID), c.Client),
		Environment: o(container.EnvironmentUUID, c.Client),
		Network:     o(s.IDtoUUID(ContainerType, container.NetworkID), c.Client),
		Stack:       o(s.IDtoUUID(StackType, container.StackID), c.Client),
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
