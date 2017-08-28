package types

import "github.com/rancher/go-rancher/v3"

type Service struct {
	EnvironmentUUID string                 `json:"environment_uuid"`
	ExternalIPs     []string               `json:"external_ips"`
	FQDN            string                 `json:"fqdn"`
	Global          bool                   `json:"global"`
	HealthCheck     *HealthcheckInfo       `json:"health_check"`
	HealthState     string                 `json:"health_state"`
	Hostname        string                 `json:"hostname"`
	InstanceIDs     []string               `json:"-"`
	Kind            string                 `json:"-"`
	Labels          map[string]string      `json:"labels"`
	LBConfig        *client.LbConfig       `json:"-"`
	Links           []Link                 `json:"-"`
	Metadata        map[string]interface{} `json:"metadata"`
	Name            string                 `json:"name"`
	Ports           []PublicEndpoint       `json:"-"`
	Scale           int64                  `json:"scale"`
	Selector        string                 `json:"selector"`
	Sidekicks       []string               `json:"sidekicks"`
	StackID         string                 `json:"-"`
	State           string                 `json:"state"`
	Token           string                 `json:"-"`
	UUID            string                 `json:"uuid"`
	VIP             string                 `json:"vip"`
}

type ServiceResponse struct {
	Service
	ServiceDynamic
}

type ServiceDynamic struct {
	Containers   []Object `json:"containers"`
	KindOutput   string   `json:"kind"`
	MetadataKind string   `json:"metadata_kind"`
	PortsOutput  []string `json:"ports"`
	StackName    string   `json:"stack_name"`
	StackUUID    string   `json:"stack_uuid"`
	TokenOutput  string   `json:"token"`

	LBConfigOutput *LBConfig              `json:"lb_config"`
	LinksOutput    map[string]interface{} `json:"links"`
}

func (s *Service) GetEnvironmentUUID() string {
	return s.EnvironmentUUID
}

func (s *Service) GetStackID() string {
	return s.StackID
}

type LBConfig struct {
	CertificateIDs       []string                            `json:"certificate_ids"`
	Config               string                              `json:"config"`
	DefaultCertificateID string                              `json:"default_certificate_id"`
	PortRules            []PortRule                          `json:"port_rules"`
	StickinessPolicy     *LoadBalancerCookieStickinessPolicy `json:"stickiness_policy"`
}

type PortRule struct {
	BackendName   string `json:"backend_name"`
	Container     string `json:"container"`
	ContainerUUID string `json:"container_uuid"`
	Path          string `json:"path"`
	Priority      int64  `json:"priority"`
	Protocol      string `json:"protocol"`
	Selector      string `json:"selector"`
	Service       string `json:"service"`
	ServiceUUID   string `json:"service_uuid"`
	SourcePort    int64  `json:"source_port"`
	TargetPort    int64  `json:"target_port"`
}

type LoadBalancerCookieStickinessPolicy struct {
	Cookie   string `json:"cookie"`
	Domain   string `json:"domain"`
	Indirect bool   `json:"indirect"`
	Mode     string `json:"mode"`
	Name     string `json:"name"`
	Nocache  bool   `json:"nocache"`
	Postonly bool   `json:"postonly"`
}
