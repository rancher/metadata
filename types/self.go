package types

import "github.com/rancher/metadata/content"

type MetadataSelf struct {
	Container   content.Object `json:"container"`
	Service     content.Object `json:"service"`
	Host        content.Object `json:"host"`
	Environment content.Object `json:"environment"`
	Network     content.Object `json:"network"`
	Stack       content.Object `json:"stack"`
}
