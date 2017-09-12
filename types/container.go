package types

import "fmt"

type Container struct {
	CreateIndex            int64                  `json:"create_index"`
	DNS                    []string               `json:"dns"`
	DNSSearch              []string               `json:"dns_search"`
	EnvironmentUUID        string                 `json:"environment_uuid"`
	ExternalID             string                 `json:"external_id"`
	HealthCheck            *HealthcheckInfo       `json:"health_check"`
	HealthCheckHosts       []HealthcheckState     `json:"-"`
	HealthState            *string                `json:"health_state"`
	HostID                 string                 `json:"-"`
	Hostname               string                 `json:"hostname"`
	Labels                 map[string]interface{} `json:"labels"`
	Links                  []Link                 `json:"-"`
	MemoryReservation      int64                  `json:"memory_reservation"`
	MilliCPUReservation    int64                  `json:"milli_cpu_reservation"`
	Name                   string                 `json:"name"`
	NetworkFromContainerID string                 `json:"-"`
	NetworkID              string                 `json:"-"`
	Ports                  []PublicEndpoint       `json:"-"`
	PrimaryIP              string                 `json:"primary_ip"`
	PrimaryMacAddress      string                 `json:"primary_mac_address"`
	ServiceID              string                 `json:"-"`
	ServiceIndex           int64                  `json:"-"`
	StackID                string                 `json:"-"`
	StartCount             int64                  `json:"start_count"`
	State                  string                 `json:"state"`
	UUID                   string                 `json:"uuid"`
}

type ContainerResponse struct {
	Container
	ContainerDynamic
}

type ContainerDynamic struct {
	EnvironmentName          string                 `json:"environment_name"`
	HealthCheckHostsOuput    []string               `json:"health_check_hosts"`
	HostUUID                 string                 `json:"host_uuid"`
	IPs                      []string               `json:"ips"`
	LinksOutput              map[string]interface{} `json:"links"`
	MetadataKind             string                 `json:"metadata_kind"`
	NetworkFromContainerUUID string                 `json:"network_from_container_uuid"`
	NetworkUUID              string                 `json:"network_uuid"`
	PortsOutput              []string               `json:"ports"`
	ServiceIndexOutput       string                 `json:"service_index"`
	ServiceUUID              string                 `json:"service_uuid"`
	ServiceName              string                 `json:"service_name"`
	StackUUID                string                 `json:"stack_uuid"`
	StackName                string                 `json:"stack_name"`
}

type HealthcheckState struct {
	HealthState string `json:"health_state"`
	HostID      string `json:"host_id"`
}

type Link struct {
	Alias string `json:"alias"`
	Name  string `json:"name"`
}

type HealthcheckInfo struct {
	HealthyThreshold    int64  `json:"healthy_threshold"`
	InitializingTimeout int64  `json:"initializing_timeout"`
	Interval            int64  `json:"interval"`
	Port                int64  `json:"port"`
	RequestLine         string `json:"request_line"`
	ResponseTimeout     int64  `json:"response_timeout"`
	UnhealthyThreshold  int64  `json:"unhealthy_threshold"`
}

type PublicEndpoint struct {
	AgentIPAddress string `json:"agent_ip_address"`
	BindAll        bool   `json:"bind_all"`
	BindIPAddress  string `json:"bind_ip_address"`
	FQDN           string `json:"fqdn"`
	HostID         string `json:"host_id"`
	InstanceID     string `json:"instance_id"`
	IPAddress      string `json:"ip_address"`
	PrivatePort    int64  `json:"private_port"`
	Protocol       string `json:"protocol"`
	PublicPort     int64  `json:"public_port"`
	ServiceID      string `json:"service_id"`
}

func (p PublicEndpoint) String() string {
	return fmt.Sprintf("%s:%d:%d/%s", p.BindIPAddress, p.PublicPort, p.PrivatePort, p.Protocol)
}

func (c *Container) GetEnvironmentUUID() string {
	return c.EnvironmentUUID
}
