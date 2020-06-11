package gcs

import (
	"fmt"
)

func j2y(v interface{}) interface{} {
	switch v.(type) {
	case map[string]interface{}:
		m := v.(map[string]interface{})
		for k, v := range m {
			m[k] = j2y(v)
		}
		return m

	case map[interface{}]interface{}:
		m := v.(map[interface{}]interface{})
		mm := make(map[string]interface{})
		for k, v := range m {
			mm[fmt.Sprintf("%s", k)] = j2y(v)
		}
		return mm

	case []interface{}:
		l := v.([]interface{})
		for i, v := range l {
			l[i] = j2y(v)
		}
		return l
	}

	return v
}
