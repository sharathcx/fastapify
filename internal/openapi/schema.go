package openapi

import (
	"reflect"
	"strings"
)

func buildSchema(t reflect.Type) map[string]interface{} {
	if t == nil {
		return map[string]interface{}{}
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() == reflect.Struct {
		if t.Name() == "Time" {
			return map[string]interface{}{"type": "string", "format": "date-time"}
		}

		props := map[string]interface{}{}
		required := []string{}

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			name := field.Tag.Get("json")
			if name == "" {
				name = field.Tag.Get("form")
			}
			if name == "" {
				name = field.Name
			}
			if name == "-" {
				continue
			}

			name = strings.Split(name, ",")[0]
			if strings.Contains(field.Tag.Get("binding"), "required") {
				required = append(required, name)
			}
			props[name] = buildSchema(field.Type)
		}

		schema := map[string]interface{}{
			"type":       "object",
			"properties": props,
		}
		if len(required) > 0 {
			schema["required"] = required
		}
		return schema
	} else if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		return map[string]interface{}{
			"type":  "array",
			"items": buildSchema(t.Elem()),
		}
	}

	valType := typeToOAS(t.Kind())
	return map[string]interface{}{"type": valType}
}

func typeToOAS(k reflect.Kind) string {
	switch k {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	default:
		return "object"
	}
}
