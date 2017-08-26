package types

type Stack struct {
	EnvironmentUUID string `json:"environment_uuid,omitempty"`
	HealthState     string `json:"health_state,omitempty"`
	Name            string `json:"name,omitempty"`
	UUID            string `json:"uuid,omitempty"`
}

func (s *Stack) GetEnvironmentUUID() string {
	return s.EnvironmentUUID
}
