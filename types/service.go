package types

type ServiceChildren struct {
	Containers []Object `json:"containers,omitempty"`
}

type Service struct {
	EnvironmentUUID string                 `json:"environment_uuid,omitempty"`
	ExternalIPs     []string               `json:"external_ips,omitempty"`
	FQDN            string                 `json:"fqdn,omitempty"`
	Global          bool                   `json:"global,omitempty"`
	HealthState     string                 `json:"health_state,omitempty"`
	Hostname        string                 `json:"hostname,omitempty"`
	Instances       interface{}            `json:"instances,omitempty"`
	Kind            string                 `json:"kind,omitempty"`
	Labels          map[string]interface{} `json:"labels,omitempty"`
	Links           map[string]interface{} `json:"links,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Name            string                 `json:"name,omitempty"`
	Ports           []PublicEndpoint       `json:"ports,omitempty"`
	Scale           int64                  `json:"scale,omitempty"`
	Selector        string                 `json:"selector,omitempty"`
	Sidekicks       []string               `json:"sidekicks,omitempty"`
	StackID         string                 `json:"stack_id,omitempty"`
	State           string                 `json:"state,omitempty"`
	Token           string                 `json:"token,omitempty"`
	UUID            string                 `json:"uuid,omitempty"`
	VIP             string                 `json:"vip,omitempty"`

	ServiceChildren
}

func (s *Service) GetEnvironmentUUID() string {
	return s.EnvironmentUUID
}
