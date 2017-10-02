package convert

import (
	"fmt"
	"strings"

	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/metadata/content"
	"github.com/rancher/metadata/types"
)

type ServiceWrapper struct {
	Client       content.Client
	Service      *client.ServiceInfo
	Store        content.Store
	IncludeToken bool
}

func NewServiceObject(obj interface{}, c content.Client, store content.Store) content.Object {
	return &WrappedObject{
		Wrapped: &ServiceWrapper{
			Client:  c,
			Service: obj.(*client.ServiceInfo),
			Store:   store,
		},
	}
}

func (c *ServiceWrapper) wrapped() interface{} {
	result := &types.ServiceResponse{
		EnvironmentUUID: c.Service.EnvironmentUuid,
		ExternalIPs:     c.Service.ExternalIps,
		FQDN:            c.Service.Fqdn,
		Global:          c.Service.Global,
		HealthState:     c.Service.HealthState,
		Hostname:        c.Service.Hostname,
		Kind:            c.Service.Kind,
		Labels:          c.Service.Labels,
		Metadata:        c.Service.Metadata,
		Scale:           c.Service.Scale,
		Selector:        c.Service.Selector,
		State:           c.Service.State,
		Token:           c.Service.Token,
		UUID:            c.Service.Uuid,
		VIP:             c.Service.Vip,
		Expose:          c.Service.Expose,
		MetadataKind:    "service",
	}

	if result.Kind == "scalingGroup" {
		result.Kind = "service"
	}

	if c.Client.Version == content.V1 {
		result.Name = c.Service.Name
		result.Sidekicks = c.Service.Sidekicks
	} else {
		result.Name = strings.ToLower(c.Service.Name)
		if c.Service.Sidekicks != nil {
			var lowercased []string
			for _, value := range c.Service.Sidekicks {
				lowercased = append(lowercased, strings.ToLower(value))
			}
			result.Sidekicks = lowercased
		} else {
			result.Sidekicks = []string{}
		}
	}

	if c.Service.HealthCheck.Interval != 0 {
		result.HealthCheck = &types.HealthcheckInfo{
			HealthyThreshold:    c.Service.HealthCheck.HealthyThreshold,
			InitializingTimeout: c.Service.HealthCheck.InitializingTimeout,
			Interval:            c.Service.HealthCheck.Interval,
			Port:                c.Service.HealthCheck.Port,
			RequestLine:         c.Service.HealthCheck.RequestLine,
			ResponseTimeout:     c.Service.HealthCheck.ResponseTimeout,
			UnhealthyThreshold:  c.Service.HealthCheck.UnhealthyThreshold,
		}
	}

	result.Containers = []content.Object{}
	for _, containerID := range c.Service.InstanceIds {
		container := c.Store.ContainerByID(containerID)
		if container != nil {
			result.Containers = append(result.Containers, NewContainerObject(container, c.Client, c.Store))
		}
	}

	result.Ports = []string{}
	for _, port := range c.Service.Ports {
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
		result.Ports = append(result.Ports, portString)
	}

	stack := c.Store.StackByID(c.Service.StackId)
	if stack != nil {
		result.StackUUID = stack.Uuid
		if c.Client.Version == content.V1 {
			result.StackName = stack.Name
		} else {
			result.StackName = strings.ToLower(stack.Name)
		}
	}

	if !c.IncludeToken {
		result.Token = ""
	}

	result.LBConfig = generateLBConfig(result, c.Service, c.Store)
	result.Links = c.resolveServiceLinks(result, c.Service, c.Store)

	env := c.Store.EnvironmentByUUID(result.EnvironmentUUID)
	if env != nil {
		result.EnvironmentName = env.Name
	}

	return result
}

func generateLBConfig(response *types.ServiceResponse, service *client.ServiceInfo, store content.Store) *types.LBConfig {
	if service.LbConfig == nil {
		return nil
	}

	result := &types.LBConfig{
		CertificateIDs:       service.LbConfig.CertificateIds,
		Config:               service.LbConfig.Config,
		DefaultCertificateID: service.LbConfig.DefaultCertificateId,
	}

	if service.LbConfig.StickinessPolicy != nil {
		result.StickinessPolicy = &types.LoadBalancerCookieStickinessPolicy{
			Cookie:   service.LbConfig.StickinessPolicy.Cookie,
			Domain:   service.LbConfig.StickinessPolicy.Domain,
			Indirect: service.LbConfig.StickinessPolicy.Indirect,
			Mode:     service.LbConfig.StickinessPolicy.Mode,
			Name:     service.LbConfig.StickinessPolicy.Name,
			Nocache:  service.LbConfig.StickinessPolicy.Nocache,
			Postonly: service.LbConfig.StickinessPolicy.Postonly,
		}
	}
	var ports []types.PortRule
	for _, rule := range service.LbConfig.PortRules {
		newRule := types.PortRule{
			BackendName: rule.BackendName,
			Path:        rule.Path,
			Priority:    rule.Priority,
			Protocol:    rule.Protocol,
			Selector:    rule.Selector,
			SourcePort:  rule.SourcePort,
			TargetPort:  rule.TargetPort,
			Hostname:    rule.Hostname,
		}

		if rule.InstanceId != "" {
			container := store.ContainerByID(rule.InstanceId)
			if container != nil {
				newRule.Container = fmt.Sprintf("%s/%s", response.StackName, container.Name)
				newRule.ContainerUUID = container.Uuid
			}
		}

		if rule.ServiceId != "" {
			target := store.ServiceByID(rule.ServiceId)
			if target != nil {
				targetStack := store.StackByID(target.StackId)
				if targetStack == nil {
					continue
				}
				newRule.Service = fmt.Sprintf("%s/%s", targetStack.Name, target.Name)
				newRule.ServiceUUID = target.Uuid
			}
		}
		ports = append(ports, newRule)
	}
	result.PortRules = ports
	return result
}

func (c *ServiceWrapper) resolveServiceLinks(response *types.ServiceResponse, service *client.ServiceInfo, store content.Store) map[string]interface{} {
	result := map[string]interface{}{}

	for _, link := range service.Links {
		alias := link.Alias
		if alias == "" {
			alias = link.Name
		}

		if c.Client.Version == content.V1 {
			result[link.Name] = alias
		} else if c.Client.Version == content.V2 {
			result[strings.ToLower(link.Name)] = strings.ToLower(alias)
		} else {
			stackName := response.StackName
			serviceName := link.Name
			parts := strings.SplitN(link.Name, "/", 2)
			if len(parts) == 2 {
				stackName = parts[0]
				serviceName = parts[1]
			}
			target := store.ServiceByName(service.EnvironmentUuid, stackName, serviceName)
			if target == nil {
				result[strings.ToLower(alias)] = nil
			} else {
				result[strings.ToLower(alias)] = target.Uuid
			}
		}

	}
	return result
}
