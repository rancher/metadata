package types

type HostResponse struct {
	AgentIP         string            `json:"agent_ip"`
	AgentState      string            `json:"agent_state"`
	EnvironmentUUID string            `json:"environment_uuid"`
	Hostname        string            `json:"hostname"`
	Labels          map[string]string `json:"labels"`
	Memory          int64             `json:"memory"`
	MilliCPU        int64             `json:"milli_cpu"`
	Name            string            `json:"name"`
	State           string            `json:"state"`
	UUID            string            `json:"uuid"`

	MetadataKind string `json:"metadata_kind"`
}
