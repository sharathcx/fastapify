package fastapify

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

func (w *Wrapper) SetupSwagger(jsonPath string) {
	w.Engine.GET(jsonPath, func(c *gin.Context) {
		docs := BuildOpenAPI(w.Routes)
		c.JSON(http.StatusOK, docs)
	})

	docsHTML := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Fastapify API</title>
    <style>
        body { margin: 0; }
    </style>
</head>
<body>
    <script id="api-reference" data-url="` + jsonPath + `" data-configuration='{"theme":"deepSpace","layout":"modern","hideModels":true}'></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`

	w.Engine.GET("/docs", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(docsHTML))
	})
}

func BuildOpenAPI(routes []RouteMeta) map[string]interface{} {
	docs := map[string]interface{}{
		"openapi": "3.0.3",
		"info": map[string]interface{}{
			"title":   "Auto-Generated API",
			"version": "1.0.0",
		},
		"paths": map[string]interface{}{},
		"components": map[string]interface{}{
			"securitySchemes": map[string]interface{}{
				"BearerAuth": map[string]interface{}{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
				},
			},
		},
		"security": []map[string]interface{}{
			{"BearerAuth": []string{}},
		},
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
							"schema": func() map[string]interface{} {
								if route.ResponseType == nil {
									return map[string]interface{}{"type": "object"}
								}
								return buildSchema(route.ResponseType)
							}(),
						},
					},
				},
				"400": map[string]interface{}{"description": "Bad Request or Validation Error"},
			},
		}

		// Path params — use ParamsType schema if available, otherwise extract from path
		params := []interface{}{}
		if route.ParamsType != nil && route.ParamsType.Kind() == reflect.Struct {
			for i := 0; i < route.ParamsType.NumField(); i++ {
				field := route.ParamsType.Field(i)
				uriTag := field.Tag.Get("uri")
				if uriTag == "" {
					continue
				}
				required := strings.Contains(field.Tag.Get("binding"), "required")
				params = append(params, map[string]interface{}{
					"name":     uriTag,
					"in":       "path",
					"required": required,
					"schema":   map[string]string{"type": typeToOAS(field.Type.Kind())},
				})
			}
		} else {
			pathParamNames := extractParamNames(route.Path)
			for _, name := range pathParamNames {
				params = append(params, map[string]interface{}{
					"name":     name,
					"in":       "path",
					"required": true,
					"schema":   map[string]string{"type": "string"},
				})
			}
		}

		// Query params from BodyType's `form` tags (for GET requests)
		if route.BodyType != nil && route.BodyType.Kind() == reflect.Struct {
			for i := 0; i < route.BodyType.NumField(); i++ {
				field := route.BodyType.Field(i)
				formTag := field.Tag.Get("form")
				required := strings.Contains(field.Tag.Get("binding"), "required")

				if formTag != "" {
					params = append(params, map[string]interface{}{
						"name":     formTag,
						"in":       "query",
						"required": required,
						"schema":   map[string]string{"type": typeToOAS(field.Type.Kind())},
					})
				}
			}
		}

		if len(params) > 0 {
			methodObj["parameters"] = params
		}

		// Request body for write methods
		if route.BodyType != nil && (route.Method == "POST" || route.Method == "PUT" || route.Method == "PATCH") {
			schema := buildSchema(route.BodyType)
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

		pathItem := pathObj[swaggerPath].(map[string]interface{})
		pathItem[strings.ToLower(route.Method)] = methodObj
	}

	return docs
}

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
