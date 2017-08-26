package content

import (
	"encoding/json"
	"reflect"
)

type wrapped interface {
	wrapped() interface{}
}

type WrappedObject struct {
	Wrapped wrapped
}

func (w *WrappedObject) Get(key string) (interface{}, bool) {
	return GetValue(w.Wrapped.wrapped(), key)
}

func (w *WrappedObject) Map() (map[string]interface{}, error) {
	bytes, err := w.MarshalJSON()
	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{}
	return result, json.Unmarshal(bytes, &result)
}

func (w *WrappedObject) Name() string {
	obj := w.Wrapped.wrapped()
	v := reflect.ValueOf(obj)
	if v.Type().Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	if t.Kind() == reflect.Struct {
		f, ok := t.FieldByName("Name")
		if ok {
			return v.FieldByIndex(f.Index).String()
		}
	}
	return ""
}

func (w *WrappedObject) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.Wrapped.wrapped())
}
