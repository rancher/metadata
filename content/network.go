package content

import (
	"github.com/rancher/metadata/types"
)

type NetworkWrapper struct {
	Client  Client
	Network *types.Network
}

func NewNetworkObject(obj interface{}, client Client, store Store) types.Object {
	return &WrappedObject{
		Wrapped: &NetworkWrapper{
			Client:  client,
			Network: obj.(*types.Network),
		},
	}
}

func (c *NetworkWrapper) wrapped() interface{} {
	return &types.NetworkResponse{
		Network: *c.Network,
		NetworkDynamic: types.NetworkDynamic{
			MetadataKind: "network",
		},
	}
}
