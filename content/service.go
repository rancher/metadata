package content

import (
	"fmt"
	"strings"

	"github.com/rancher/metadata/types"
)

type ServiceWrapper struct {
	Client       Client
	Service      *types.Service
	Store        Store
	IncludeToken bool
}

func NewServiceObject(obj interface{}, client Client, store Store) types.Object {
	return &WrappedObject{
		Wrapped: &ServiceWrapper{
			Client:  client,
			Service: obj.(*types.Service),
			Store:   store,
		},
	}
}

func (c *ServiceWrapper) wrapped() interface{} {
	result := &types.ServiceResponse{
		Service: *c.Service,
		ServiceDynamic: types.ServiceDynamic{
			MetadataKind: "service",
		},
	}
	result.Name = strings.ToLower(result.Name)

	result.KindOutput = result.Kind
	if result.Kind == "scalingGroup" {
		result.KindOutput = "service"
	}

	result.Containers = []types.Object{}
	for _, containerID := range result.InstanceIDs {
		container := c.Store.ContainerByID(containerID)
		if container != nil {
			result.Containers = append(result.Containers, NewContainerObject(container, c.Client, c.Store))
		}
	}

	result.PortsOutput = []string{}
	for _, port := range result.Ports {
		result.PortsOutput = append(result.PortsOutput, port.String())
	}

	stack := c.Store.StackByID(result.StackID)
	if stack != nil {
		result.StackUUID = stack.UUID
		result.StackName = strings.ToLower(stack.Name)
	}

	if c.IncludeToken {
		result.TokenOutput = result.Token
	}

	result.LBConfigOutput = generateLBConfig(result, c.Store)
	result.LinksOutput = resolveServiceLinks(result, c.Store)

	return result
}

func generateLBConfig(service *types.ServiceResponse, store Store) *types.LBConfig {
	if service.LBConfig == nil {
		return nil
	}

	result := &types.LBConfig{
		CertificateIDs:       service.LBConfig.CertificateIds,
		Config:               service.LBConfig.Config,
		DefaultCertificateID: service.LBConfig.DefaultCertificateId,
	}

	if service.LBConfig.StickinessPolicy != nil {
		result.StickinessPolicy = &types.LoadBalancerCookieStickinessPolicy{
			Cookie:   service.LBConfig.StickinessPolicy.Cookie,
			Domain:   service.LBConfig.StickinessPolicy.Domain,
			Indirect: service.LBConfig.StickinessPolicy.Indirect,
			Mode:     service.LBConfig.StickinessPolicy.Mode,
			Name:     service.LBConfig.StickinessPolicy.Name,
			Nocache:  service.LBConfig.StickinessPolicy.Nocache,
			Postonly: service.LBConfig.StickinessPolicy.Postonly,
		}
	}

	for _, rule := range service.LBConfig.PortRules {
		newRule := types.PortRule{
			BackendName: rule.BackendName,
			Path:        rule.Path,
			Priority:    rule.Priority,
			Protocol:    rule.Protocol,
			Selector:    rule.Selector,
			SourcePort:  rule.SourcePort,
			TargetPort:  rule.TargetPort,
		}

		if rule.InstanceId != "" {
			container := store.ContainerByID(rule.InstanceId)
			if container != nil {
				newRule.Container = fmt.Sprintf("%s/%s", service.StackName, container.Name)
				newRule.ContainerUUID = container.UUID
			}
		}

		if rule.ServiceId != "" {
			target := store.ServiceByID(rule.ServiceId)
			if target != nil {
				newRule.Service = fmt.Sprintf("%s/%s", service.StackName, target.Name)
				newRule.ServiceUUID = target.UUID
			}
		}
	}

	return result
}

func resolveServiceLinks(service *types.ServiceResponse, store Store) map[string]interface{} {
	result := map[string]interface{}{}

	for _, link := range service.Links {
		alias := link.Alias
		if alias == "" {
			alias = link.Name
		}

		stackName := service.StackName
		containerName := link.Name
		parts := strings.SplitN(link.Name, "/", 2)
		if len(parts) == 2 {
			stackName = parts[0]
			containerName = parts[1]
		}

		target := store.ServiceByName(service.EnvironmentUUID, stackName, containerName)
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
