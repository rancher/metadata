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
		UUID:            c.Stack.Uuid,
		MetadataKind:    "stack",
	}

	env := c.Store.EnvironmentByUUID(result.EnvironmentUUID)
	if env != nil {
		result.EnvironmentName = env.Name
	}

	if c.Client.Version == content.V1 || c.Client.Version == content.V2 {
		result.Name = c.Stack.Name
	} else {
		result.Name = strings.ToLower(c.Stack.Name)
	}

	if c.Client.Version == content.V1 {
		resultVersioned := &types.StackResponseV1{
			StackResponse: result,
			Services:      []string{},
		}

		for _, svc := range c.Store.ByStack(content.ServiceType, c.Client, result.UUID) {
			resultVersioned.Services = append(resultVersioned.Services, svc.Name())
		}

		return resultVersioned
	}

	resultVersioned := &types.StackResponseV2V3V4{
		StackResponse: result,
		Services:      c.Store.ByStack(content.ServiceType, c.Client, result.UUID),
	}
	return resultVersioned
}
