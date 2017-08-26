package content

import (
	"github.com/rancher/metadata/types"
)

type Container struct {
	Client    Client
	Container *types.Container
}

func NewContainerObject(obj interface{}, client Client, store Store) types.Object {
	return &WrappedObject{
		Wrapped: &Container{
			Client:    client,
			Container: obj.(*types.Container),
		},
	}
}

func (c *Container) wrapped() interface{} {
	//switch c.Client.Version {
	//case V3:
	return c.Container
	//}
	//
	//return nil
}
