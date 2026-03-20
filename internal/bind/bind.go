package bind

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"

	"github.com/sharathcx/fastapify/internal/response"
)

// Bind validates and binds the request into the given struct.
// Automatically:
//  1. Binds URI parameters (and makes them immutable)
//  2. Detects HTTP method for Body vs Query binding
//  3. Sends error response on failure
//
// Returns true on success, false on failure.
func Bind(c *gin.Context, req any) bool {
	// Step 1: Bind URI params and save their values for protection
	uriValues := make(map[int]reflect.Value)
	reqVal := reflect.ValueOf(req).Elem()
	if reqVal.Kind() == reflect.Struct {
		_ = c.ShouldBindUri(req)

		// Snapshot URI-tagged fields
		reqType := reqVal.Type()
		for i := 0; i < reqType.NumField(); i++ {
			if reqType.Field(i).Tag.Get("uri") != "" {
				uriValues[i] = reflect.ValueOf(reqVal.Field(i).Interface())
			}
		}
	}

	// Step 2: Bind Body or Query
	var err error
	switch c.Request.Method {
	case http.MethodGet, http.MethodDelete:
		err = c.ShouldBindQuery(req)
	default:
		err = c.ShouldBindJSON(req)
	}

	if err != nil && err.Error() != "EOF" {
		statusCode, res := response.HandleError(err)
		c.JSON(statusCode, res)
		return false
	}

	// Step 3: Restore URI values (Protection against body override)
	for i, val := range uriValues {
		reqVal.Field(i).Set(val)
	}

	return true
}
