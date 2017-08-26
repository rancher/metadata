package server

import (
	"encoding/json"
	"fmt"
	"github.com/rancher/metadata/types"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strings"
)

func respondError(w http.ResponseWriter, req *http.Request, msg string, statusCode int) {
	obj := map[string]interface{}{
		"message": msg,
		"type":    "error",
		"code":    statusCode,
	}

	switch contentType(req) {
	case ContentText:
		http.Error(w, msg, statusCode)
	case ContentJSON:
		bytes, err := json.Marshal(obj)
		if err == nil {
			http.Error(w, string(bytes), statusCode)
		} else {
			http.Error(w, "{\"type\": \"error\", \"message\": \"JSON marshal error\"}", http.StatusInternalServerError)
		}
	}
}

func respondSuccess(w http.ResponseWriter, req *http.Request, val interface{}) {
	switch contentType(req) {
	case ContentText:
		respondText(w, req, val)
	case ContentJSON:
		respondJSON(w, req, val)
	}
}

func respondText(w http.ResponseWriter, req *http.Request, val interface{}) {
	if val == nil {
		fmt.Fprint(w, "")
		return
	}

	if obj, ok := val.(types.Object); ok {
		mapObj, err := obj.Map()
		if err != nil {
			respondError(w, req, err.Error(), 500)
			return
		}
		val = mapObj
	}

	switch v := val.(type) {
	case float64:
		// The default format has extra trailing zeros
		str := strings.TrimRight(fmt.Sprintf("%f", v), "0")
		str = strings.TrimRight(str, ".")
		fmt.Fprint(w, str)
	case map[string]interface{}:
		out := make([]string, len(v))
		i := 0
		for k, vv := range v {
			t := reflect.ValueOf(vv)
			if t.Kind() == reflect.Map || t.Kind() == reflect.Slice {
				out[i] = fmt.Sprintf("%s/\n", url.QueryEscape(k))
			} else {
				out[i] = fmt.Sprintf("%s\n", url.QueryEscape(k))
			}
			i++
		}

		sort.Strings(out)
		for _, vv := range out {
			fmt.Fprint(w, vv)
		}
	default:
		if reflect.TypeOf(v).Kind() == reflect.Slice {
			sliceValue := reflect.ValueOf(v)
			for k := 0; k < sliceValue.Len(); k++ {
				vv := sliceValue.Index(k).Interface()
				kind := reflect.TypeOf(vv).Kind()
				needsSlash := kind == reflect.Map || kind == reflect.Slice
				name := getName(vv)

				if name != "" {
					fmt.Fprintf(w, "%d=%s\n", k, url.QueryEscape(name))
				} else if needsSlash {
					// If the child is a map or array, show index ("0/")
					fmt.Fprintf(w, "%d/\n", k)
				} else {
					// Otherwise, show index ("0" )
					fmt.Fprintf(w, "%d\n", k)
				}
			}
		} else {
			fmt.Fprint(w, v)
		}
	}
}

func getName(obj interface{}) string {
	if named, ok := obj.(interface {
		Name() string
	}); ok {
		return named.Name()
	}
	return ""
}

func respondJSON(w http.ResponseWriter, req *http.Request, val interface{}) {
	if err := json.NewEncoder(w).Encode(val); err != nil {
		respondError(w, req, "Error serializing to JSON: "+err.Error(), http.StatusInternalServerError)
	}
}
