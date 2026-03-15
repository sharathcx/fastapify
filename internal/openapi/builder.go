package openapi

import (
	"reflect"
	"strings"

	"github.com/sharathcx/fastapify/internal/router"
)

func BuildOpenAPI(routes []router.RouteMeta) map[string]interface{} {
	docs := map[string]interface{}{
		"openapi": "3.0.3",
		"info": map[string]interface{}{
			"title":   "MagicStream Auto-Generated API",
			"version": "1.0.0",
		},
		"paths": map[string]interface{}{},
	}

	for _, route := range routes {
		pathObj := docs["paths"].(map[string]interface{})
		swaggerPath := route.Path

		if _, ok := pathObj[swaggerPath]; !ok {
			pathObj[swaggerPath] = map[string]interface{}{}
		}

		methodObj := map[string]interface{}{
			"summary": route.Method + " " + route.Path,
			"tags":    []string{route.Tag},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Success",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": buildSchema(route.Output),
						},
					},
				},
				"400": map[string]interface{}{"description": "Bad Request or Validation Error"},
			},
		}

		params := []interface{}{}
		if route.Input != nil && route.Input.Kind() == reflect.Struct {
			for i := 0; i < route.Input.NumField(); i++ {
				field := route.Input.Field(i)
				uriTag := field.Tag.Get("uri")
				formTag := field.Tag.Get("form")
				required := strings.Contains(field.Tag.Get("binding"), "required")

				if uriTag != "" {
					params = append(params, map[string]interface{}{
						"name":     uriTag,
						"in":       "path",
						"required": true,
						"schema":   map[string]string{"type": typeToOAS(field.Type.Kind())},
					})
				} else if formTag != "" && field.Tag.Get("json") == "" {
					params = append(params, map[string]interface{}{
						"name":     formTag,
						"in":       "query",
						"required": required,
						"schema":   map[string]string{"type": typeToOAS(field.Type.Kind())},
					})
				}
			}

			if route.Method == "POST" || route.Method == "PUT" || route.Method == "PATCH" {
				schema := buildSchema(route.Input)
				// Ensure it's not totally empty to avoid UI errors, or only add if properties exist.
				if props, ok := schema["properties"]; ok && len(props.(map[string]interface{})) > 0 {
					methodObj["requestBody"] = map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": schema,
							},
						},
					}
				}
			}
		}

		if len(params) > 0 {
			methodObj["parameters"] = params
		}

		pathItem := pathObj[swaggerPath].(map[string]interface{})
		pathItem[strings.ToLower(route.Method)] = methodObj
	}

	return docs
}
