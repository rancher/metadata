package types

type EnvironmentChildren struct {
	Containers []Object `json:"containers"`
	Services   []Object `json:"services"`
	Networks   []Object `json:"networks"`
	Hosts      []Object `json:"hosts"`
	Stacks     []Object `json:"stacks"`
}

type Environment struct {
	ExternalID string `json:"external_id,omitempty"`
	System     bool   `json:"system,omitempty"`
	UUID       string `json:"uuid,omitempty"`
	Version    string `json:"version,omitempty"`

	EnvironmentChildren
}
