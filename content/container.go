package content

import (
	"strconv"
	"strings"

	"github.com/rancher/metadata/types"
)

type ContainerWrapper struct {
	Client    Client
	Container *types.Container
	Store     Store
}

func NewContainerObject(obj interface{}, client Client, store Store) types.Object {
	return &WrappedObject{
		Wrapped: &ContainerWrapper{
			Client:    client,
			Container: obj.(*types.Container),
			Store:     store,
		},
	}
}

func (c *ContainerWrapper) wrapped() interface{} {
	container := types.ContainerResponse{
		Container: *c.Container,
		ContainerDynamic: types.ContainerDynamic{
			MetadataKind: "container",
		},
	}
	container.HostUUID = c.Store.IDtoUUID(HostType, container.HostID)
	container.NetworkFromContainerUUID = c.Store.IDtoUUID(ContainerType, container.NetworkFromContainerID)
	container.NetworkUUID = c.Store.IDtoUUID(NetworkType, container.NetworkID)
	container.ServiceIndexOutput = strconv.FormatInt(container.ServiceIndex, 10)
	container.HealthCheckHostsOuput = []string{} // don't want the output to be nil

	for _, info := range container.HealthCheckHosts {
		container.HealthCheckHostsOuput = append(container.HealthCheckHostsOuput,
			c.Store.IDtoUUID(HostType, info.HostID))
	}

	setupNetworking(&container, c.Store)

	service := c.Store.ServiceByID(container.ServiceID)
	if service != nil {
		container.ServiceUUID = service.UUID
		container.ServiceName = service.Name
	}

	stack := c.Store.StackByID(container.StackID)
	if stack != nil {
		container.StackUUID = stack.UUID
		container.StackName = stack.Name
	}

	for _, port := range container.Ports {
		container.PortsOutput = append(container.PortsOutput, port.String())
	}

	container.LinksOutput = resolveContainerLinks(&container, c.Store)
	return &container
}

func setupNetworking(container *types.ContainerResponse, store Store) {
	network := store.NetworkByID(container.NetworkID)
	if network != nil && network.Kind == "host" {
		host := store.HostByID(container.HostID)
		if host != nil {
			container.PrimaryIP = host.AgentIP
			container.PrimaryMacAddress = ""
		}
	} else if network != nil && network.Kind == "container" {
		netContainer := store.ContainerByID(container.NetworkFromContainerID)
		if netContainer != nil {
			container.PrimaryIP = netContainer.PrimaryIP
			container.PrimaryMacAddress = netContainer.PrimaryMacAddress
		}
	}

	if container.PrimaryIP != "" {
		container.IPs = []string{container.PrimaryIP}
	}
}

func resolveContainerLinks(container *types.ContainerResponse, store Store) map[string]interface{} {
	result := map[string]interface{}{}

	for _, link := range container.Links {
		alias := link.Alias
		if alias == "" {
			alias = link.Name
		}

		stackName := container.StackName
		containerName := link.Name
		parts := strings.SplitN(link.Name, "/", 2)
		if len(parts) == 2 {
			stackName = parts[0]
			containerName = parts[1]
		}

		target := store.ContainerByName(container.EnvironmentUUID, stackName, containerName)
		if target == nil {
			result[alias] = nil
		} else {
			result[alias] = target.UUID
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}
