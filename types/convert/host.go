package convert

import (
	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/metadata/content"
	"github.com/rancher/metadata/types"
)

type HostWrapper struct {
	Client content.Client
	Host   *client.HostInfo
}

func NewHostObject(obj interface{}, c content.Client, store content.Store) content.Object {
	return &WrappedObject{
		Wrapped: &HostWrapper{
			Client: c,
			Host:   obj.(*client.HostInfo),
		},
	}
}

func (c *HostWrapper) wrapped() interface{} {
	name := c.Host.Name
	if name == "" {
		name = c.Host.Hostname
	}
	return &types.HostResponse{
		AgentIP:         c.Host.AgentIp,
		AgentState:      c.Host.AgentState,
		EnvironmentUUID: c.Host.EnvironmentUuid,
		Hostname:        c.Host.Hostname,
		Labels:          c.Host.Labels,
		Memory:          c.Host.Memory,
		MilliCPU:        c.Host.MilliCpu,
		Name:            name,
		State:           c.Host.State,
		UUID:            c.Host.Uuid,
		MetadataKind:    "host",
	}
}
