package content

import (
	"strings"

	"github.com/rancher/metadata/types"
)

type Stack struct {
	Client Client
	Stack  *types.Stack
	Store  Store
}

func NewStackObject(obj interface{}, client Client, store Store) types.Object {
	return &WrappedObject{
		Wrapped: &Stack{
			Client: client,
			Stack:  obj.(*types.Stack),
			Store:  store,
		},
	}
}

func (c *Stack) wrapped() interface{} {
	result := &types.StackResponse{
		Stack: *c.Stack,
		StackDynamic: types.StackDynamic{
			MetadataKind: "stack",
		},
	}
	result.Name = strings.ToLower(result.Name)

	env := c.Store.EnvironmentByUUID(result.EnvironmentUUID)
	if env != nil {
		result.EnvironmentName = env.Name
	}

	result.Services = c.Store.ByStack(ServiceType, c.Client, result.UUID)
	return c.Stack
}
