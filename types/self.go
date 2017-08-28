package types

type MetadataSelf struct {
	Container   Object `json:"container"`
	Service     Object `json:"service"`
	Host        Object `json:"host"`
	Environment Object `json:"environment"`
	Network     Object `json:"network"`
	Stack       Object `json:"stack"`
}
