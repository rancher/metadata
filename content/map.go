package content

import (
	"reflect"
	"strings"
)

func GetValue(obj interface{}, key string) (interface{}, bool) {
	v := reflect.ValueOf(obj)
	if v.Type().Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	if t.Kind() != reflect.Struct {
		return nil, false
	}

	f, ok := t.FieldByName(key)
	if !ok {
		f, ok = t.FieldByName(strings.ToLower(key))
	}
	if !ok {
		f, ok = t.FieldByName(toCamelCase(key))
	}
	if !ok {
		return nil, false
	}
	return v.FieldByIndex(f.Index).Interface(), true
}

func toCamelCase(val string) string {
	parts := strings.Split(val, "_")
	result := make([]string, len(parts))

	for i, part := range parts {
		if len(part) > 0 {
			result[i] = strings.ToUpper(part[0:1]) + part[1:]
		}
	}

	return strings.Join(result, "")
}
