package convert

import (
	"strings"

	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/metadata/content"
	"github.com/rancher/metadata/types"
)

type Stack struct {
	Client content.Client
	Stack  *client.StackInfo
	Store  content.Store
}

func NewStackObject(obj interface{}, c content.Client, store content.Store) content.Object {
	return &WrappedObject{
		Wrapped: &Stack{
			Client: c,
			Stack:  obj.(*client.StackInfo),
			Store:  store,
		},
	}
}

func (c *Stack) wrapped() interface{} {
	result := &types.StackResponse{
		ID:              c.Stack.Id,
		EnvironmentUUID: c.Stack.EnvironmentUuid,
		HealthState:     c.Stack.HealthState,
		Name:            strings.ToLower(c.Stack.Name),
		UUID:            c.Stack.Uuid,
		MetadataKind:    "stack",
	}

	env := c.Store.EnvironmentByUUID(result.EnvironmentUUID)
	if env != nil {
		result.EnvironmentName = env.Name
	}

	result.Services = c.Store.ByStack(content.ServiceType, c.Client, result.UUID)
	return result
}
