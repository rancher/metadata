package types

import "github.com/rancher/metadata/content"

type EnvironmentResponse struct {
	Name       string `json:"name"`
	ExternalID string `json:"external_id"`
	System     bool   `json:"system"`
	UUID       string `json:"uuid"`

	Containers []content.Object `json:"containers"`
	Services   []content.Object `json:"services"`
	Networks   []content.Object `json:"networks"`
	Hosts      []content.Object `json:"hosts"`
	Stacks     []content.Object `json:"stacks"`
	Version    string           `json:"version"`
}
