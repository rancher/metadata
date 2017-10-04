package convert

import (
	"strconv"
	"strings"

	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/metadata/content"
	"github.com/rancher/metadata/types"
)

type ContainerWrapper struct {
	Client    content.Client
	Container *client.InstanceInfo
	Store     content.Store
}

func NewContainerObject(obj interface{}, c content.Client, store content.Store) content.Object {
	return &WrappedObject{
		Wrapped: &ContainerWrapper{
			Client:    c,
			Container: obj.(*client.InstanceInfo),
			Store:     store,
		},
	}
}

func (c *ContainerWrapper) wrapped() interface{} {
	container := types.ContainerResponse{
		CreateIndex:         c.Container.CreateIndex,
		DNS:                 c.Container.Dns,
		DNSSearch:           c.Container.DnsSearch,
		EnvironmentUUID:     c.Container.EnvironmentUuid,
		ExternalID:          c.Container.ExternalId,
		Hostname:            c.Container.Hostname,
		Labels:              c.Container.Labels,
		MemoryReservation:   c.Container.MemoryReservation,
		MilliCPUReservation: c.Container.MilliCpuReservation,
		PrimaryIP:           c.Container.PrimaryIp,
		PrimaryMacAddress:   c.Container.PrimaryMacAddress,
		StartCount:          c.Container.StartCount,
		State:               c.Container.State,
		UUID:                c.Container.Uuid,
		Expose:              c.Container.Expose,
		MetadataKind:        "container",
	}

	if c.Container.HealthState != "" {
		container.HealthState = &c.Container.HealthState
	}

	if c.Client.Version == content.V1 || c.Client.Version == content.V2 {
		container.Name = c.Container.Name
	} else {
		container.Name = strings.ToLower(c.Container.Name)
	}

	if c.Container.HealthCheck.Interval != 0 {
		container.HealthCheck = &types.HealthcheckInfo{
			HealthyThreshold:    c.Container.HealthCheck.HealthyThreshold,
			InitializingTimeout: c.Container.HealthCheck.InitializingTimeout,
			Interval:            c.Container.HealthCheck.Interval,
			Port:                c.Container.HealthCheck.Port,
			RequestLine:         c.Container.HealthCheck.RequestLine,
			ResponseTimeout:     c.Container.HealthCheck.ResponseTimeout,
			UnhealthyThreshold:  c.Container.HealthCheck.UnhealthyThreshold,
		}
	}

	container.HostUUID = c.Store.IDtoUUID(content.HostType, c.Container.HostId)
	container.NetworkFromContainerUUID = c.Store.IDtoUUID(content.ContainerType, c.Container.NetworkFromContainerId)
	container.NetworkUUID = c.Store.IDtoUUID(content.NetworkType, c.Container.NetworkId)
	container.ServiceIndex = strconv.FormatInt(c.Container.ServiceIndex, 10)
	container.HealthCheckHosts = []string{} // don't want the output to be nil

	for _, info := range c.Container.HealthCheckHosts {
		container.HealthCheckHosts = append(container.HealthCheckHosts,
			c.Store.IDtoUUID(content.HostType, info.HostId))
	}

	setupNetworking(&container, c.Container, c.Store)

	service := c.Store.ServiceByID(c.Container.ServiceId)
	if service != nil {
		container.ServiceUUID = service.Uuid
		if c.Client.Version == content.V1 || c.Client.Version == content.V2 {
			container.ServiceName = service.Name
		} else {
			container.ServiceName = strings.ToLower(service.Name)
		}
	}

	stack := c.Store.StackByID(c.Container.StackId)
	if stack != nil {
		container.StackUUID = stack.Uuid
		if c.Client.Version == content.V1 || c.Client.Version == content.V2 {
			container.StackName = stack.Name
		} else {
			container.StackName = strings.ToLower(stack.Name)
		}
	}

	container.Ports = []string{}
	for _, port := range c.Container.Ports {
		portString := types.PublicEndpoint{
			AgentIPAddress: port.AgentIpAddress,
			BindAll:        port.BindAll,
			BindIPAddress:  port.BindIpAddress,
			FQDN:           port.Fqdn,
			HostID:         port.HostId,
			InstanceID:     port.InstanceId,
			IPAddress:      port.IpAddress,
			PrivatePort:    port.PrivatePort,
			Protocol:       port.Protocol,
			PublicPort:     port.PublicPort,
			ServiceID:      port.ServiceId,
		}.String()
		container.Ports = append(container.Ports, portString)
	}

	container.Links = resolveContainerLinks(&container, c.Container, c.Store)
	env := c.Store.EnvironmentByUUID(container.EnvironmentUUID)
	if env != nil {
		container.EnvironmentName = env.Name
	}
	return &container
}

func setupNetworking(response *types.ContainerResponse, container *client.InstanceInfo, store content.Store) {
	network := store.NetworkByID(container.NetworkId)
	if network != nil && network.Kind == "host" {
		host := store.HostByID(container.HostId)
		if host != nil {
			response.PrimaryIP = host.AgentIp
			response.PrimaryMacAddress = ""
		}
	} else if network != nil && network.Kind == "response" {
		netContainer := store.ContainerByID(container.NetworkFromContainerId)
		if netContainer != nil {
			response.PrimaryIP = netContainer.PrimaryIp
			response.PrimaryMacAddress = netContainer.PrimaryMacAddress
		}
	}

	response.IPs = []string{}
	if response.PrimaryIP != "" {
		response.IPs = []string{response.PrimaryIP}
	}
}

func resolveContainerLinks(response *types.ContainerResponse, container *client.InstanceInfo, store content.Store) map[string]interface{} {
	result := map[string]interface{}{}

	for _, link := range container.Links {
		alias := link.Alias
		if alias == "" {
			alias = link.Name
		}

		stackName := response.StackName
		containerName := link.Name
		parts := strings.SplitN(link.Name, "/", 2)
		if len(parts) == 2 {
			stackName = parts[0]
			containerName = parts[1]
		}

		target := store.ContainerByName(container.EnvironmentUuid, stackName, containerName)
		if target == nil {
			result[alias] = nil
		} else {
			result[alias] = target.Uuid
		}
	}

	return result
}
