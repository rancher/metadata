package types

type Network struct {
	DefaultPolicyAction string                 `json:"default_policy_action"`
	EnvironmentUUID     string                 `json:"environment_uuid"`
	HostPorts           bool                   `json:"host_ports"`
	Kind                string                 `json:"kind"`
	Metadata            map[string]interface{} `json:"metadata"`
	Name                string                 `json:"name"`
	Policy              interface{}            `json:"policy"`
	UUID                string                 `json:"uuid"`
}

type NetworkResponse struct {
	Network
	NetworkDynamic
}

type NetworkDynamic struct {
	MetadataKind string `json:"metadata_kind"`
}

func (n *Network) GetEnvironmentUUID() string {
	return n.EnvironmentUUID
}
