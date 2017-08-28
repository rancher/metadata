package content

import (
	"github.com/rancher/metadata/types"
)

type HostWrapper struct {
	Client Client
	Host   *types.Host
}

func NewHostObject(obj interface{}, client Client, store Store) types.Object {
	return &WrappedObject{
		Wrapped: &HostWrapper{
			Client: client,
			Host:   obj.(*types.Host),
		},
	}
}

func (c *HostWrapper) wrapped() interface{} {
	return &types.HostResponse{
		Host: *c.Host,
		HostDynamic: types.HostDynamic{
			MetadataKind: "host",
		},
	}
}
