package content

import (
	"github.com/rancher/metadata/types"
)

type Stack struct {
	Client Client
	Stack  *types.Stack
}

func NewStackObject(obj interface{}, client Client, store Store) types.Object {
	return &WrappedObject{
		Wrapped: &Stack{
			Client: client,
			Stack:  obj.(*types.Stack),
		},
	}
}

func (c *Stack) wrapped() interface{} {
	//switch c.Client.Version {
	//case V3:
	return c.Stack
	//}
	//
	//return nil
}
