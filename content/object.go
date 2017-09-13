package content

type Object interface {
	Get(key string) (interface{}, bool)
	Map() (map[string]interface{}, error)
	Name() string
}
