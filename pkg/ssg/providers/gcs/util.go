package gcs

import (
	"fmt"
)

func j2y(v interface{}) interface{} {
	switch x := v.(type) {
	case map[string]interface{}:
		for k, v := range x {
			x[k] = j2y(v)
		}
		return x

	case map[interface{}]interface{}:
		m := make(map[string]interface{})
		for k, v := range x {
			m[fmt.Sprintf("%s", k)] = j2y(v)
		}
		return m

	case []interface{}:
		for i, v := range x {
			x[i] = j2y(v)
		}
		return x
	}

	return v
}
