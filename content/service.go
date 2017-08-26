package content

import (
	"github.com/rancher/metadata/types"
)

type Service struct {
	Client  Client
	Service *types.Service
	Store   Store
}

func NewServiceObject(obj interface{}, client Client, store Store) types.Object {
	return &WrappedObject{
		Wrapped: &Service{
			Client:  client,
			Service: obj.(*types.Service),
			Store:   store,
		},
	}
}

func (c *Service) wrapped() interface{} {
	//switch c.Client.Version {
	//case V3:
	result := *c.Service
	result.ServiceChildren = c.children()
	return result
	//}
	//
	//return nil
}

func (c *Service) children() types.ServiceChildren {
	return types.ServiceChildren{
		Containers: c.Store.ByService(ContainerType, c.Client, c.Service.UUID),
	}
}
