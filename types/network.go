package types

type NetworkResponse struct {
	DefaultPolicyAction string                 `json:"default_policy_action"`
	EnvironmentUUID     string                 `json:"environment_uuid"`
	HostPorts           bool                   `json:"host_ports"`
	Kind                string                 `json:"kind"`
	Metadata            map[string]interface{} `json:"metadata"`
	Name                string                 `json:"name"`
	Policy              interface{}            `json:"policy"`
	UUID                string                 `json:"uuid"`
	MetadataKind        string                 `json:"metadata_kind"`
}
