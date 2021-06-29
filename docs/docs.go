package docs

import (
	_ "embed"
	"net/http"
)

//go:embed  openapi.yaml
var openapiYAML []byte

// GetSwaggerDocument servers the Swagger JSON file.
func GetSwaggerDocument(w http.ResponseWriter, r *http.Request) {
	w.Write(openapiYAML)
}
