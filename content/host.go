package content

import (
	"github.com/rancher/metadata/types"
)

type Host struct {
	Client Client
	Host   *types.Host
}

func NewHostObject(obj interface{}, client Client, store Store) types.Object {
	return &WrappedObject{
		Wrapped: &Host{
			Client: client,
			Host:   obj.(*types.Host),
		},
	}
}

func (c *Host) wrapped() interface{} {
	//switch c.Client.Version {
	//case V3:
	return c.Host
	//}
	//
	//return nil
}
