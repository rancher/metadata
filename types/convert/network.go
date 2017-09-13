package convert

import (
	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/metadata/content"
	"github.com/rancher/metadata/types"
)

type NetworkWrapper struct {
	Client  content.Client
	Network *client.NetworkInfo
}

func NewNetworkObject(obj interface{}, c content.Client, store content.Store) content.Object {
	return &WrappedObject{
		Wrapped: &NetworkWrapper{
			Client:  c,
			Network: obj.(*client.NetworkInfo),
		},
	}
}

func (c *NetworkWrapper) wrapped() interface{} {
	return &types.NetworkResponse{
		DefaultPolicyAction: c.Network.DefaultPolicyAction,
		EnvironmentUUID:     c.Network.EnvironmentUuid,
		HostPorts:           c.Network.HostPorts,
		Kind:                c.Network.Kind,
		Metadata:            c.Network.Metadata,
		Name:                c.Network.Name,
		Policy:              c.Network.Policy,
		UUID:                c.Network.Uuid,
		MetadataKind:        "network",
	}
}
