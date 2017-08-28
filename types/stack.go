package types

type Stack struct {
	ID              string `json:"-"`
	EnvironmentUUID string `json:"environment_uuid"`
	HealthState     string `json:"health_state"`
	Name            string `json:"name"`
	UUID            string `json:"uuid"`
}

type StackResponse struct {
	Stack
	StackDynamic
}

type StackDynamic struct {
	MetadataKind    string   `json:"metadata_kind"`
	EnvironmentName string   `json:"environment_name"`
	Services        []Object `json:"services"`
}

func (s *Stack) GetEnvironmentUUID() string {
	return s.EnvironmentUUID
}
