package types

type Environment struct {
	Name       string `json:"name"`
	ExternalID string `json:"external_id"`
	System     bool   `json:"system"`
	UUID       string `json:"uuid"`
}

type EnvironmentResponse struct {
	Environment
	EnvironmentDynamic
}

type EnvironmentDynamic struct {
	Containers []Object `json:"containers"`
	Services   []Object `json:"services"`
	Networks   []Object `json:"networks"`
	Hosts      []Object `json:"hosts"`
	Stacks     []Object `json:"stacks"`
	Version    string   `json:"version"`
}
