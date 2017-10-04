package types

import "github.com/rancher/metadata/content"

type StackResponse struct {
	ID              string `json:"-"`
	EnvironmentUUID string `json:"environment_uuid"`
	HealthState     string `json:"health_state"`
	Name            string `json:"name"`
	UUID            string `json:"uuid"`

	MetadataKind    string           `json:"metadata_kind"`
	EnvironmentName string           `json:"environment_name"`
	Services        []content.Object `json:"services"`
}

type StackResponseV1 struct {
	*StackResponse
	Services []string `json:"services"`
}

type StackResponseV2V3V4 struct {
	*StackResponse
	Services []content.Object `json:"services"`
}
