package content

import (
	"github.com/rancher/metadata/types"
)

type Network struct {
	Client  Client
	Network *types.Network
}

func NewNetworkObject(obj interface{}, client Client, store Store) types.Object {
	return &WrappedObject{
		Wrapped: &Network{
			Client:  client,
			Network: obj.(*types.Network),
		},
	}
}

func (c *Network) wrapped() interface{} {
	//switch c.Client.Version {
	//case V3:
	return c.Network
	//}
	//
	//return nil
}
