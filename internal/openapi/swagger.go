package openapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sharathcx/fastapify/internal/router"
)

func SetupSwagger(w *router.Wrapper, jsonPath string) {
	w.Engine.GET(jsonPath, func(c *gin.Context) {
		docs := BuildOpenAPI(w.Routes)
		c.JSON(http.StatusOK, docs)
	})

	docsHTML := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Fastapify API</title>
    <link rel="stylesheet" type="text/css" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.0.0/swagger-ui.css">
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.0.0/swagger-ui-bundle.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.0.0/swagger-ui-standalone-preset.js"></script>
    <script>
    window.onload = function() {
        const ui = SwaggerUIBundle({
            url: "` + jsonPath + `",
            dom_id: '#swagger-ui',
            presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
            layout: "StandaloneLayout"
        })
    }
    </script>
</body>
</html>`

	w.Engine.GET("/docs", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(docsHTML))
	})
}
