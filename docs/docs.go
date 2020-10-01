package docs

//go:generate go run generator.go

import (
	"net/http"
)

var openapiYAML []byte

// GetSwaggerDocument servers the Swagger JSON file.
func GetSwaggerDocument(w http.ResponseWriter, r *http.Request) {
	w.Write(openapiYAML)
}
