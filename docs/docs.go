// Package docs Presence API.
//
// Documentation of the Presence API Web Service.
//
//     Schemes: http, https
//     Host: localhost:8080
//     BasePath: /api
//     Version: 0.0.1
//     License: MIT http://opensource.org/licenses/MIT
//     Contact: Vincent Touchard <touchardv@gmail.com>
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
// swagger:meta
package docs

import (
	"net/http"
	"os"
	"path/filepath"
)

// GetSwaggerDocument servers the Swagger JSON file.
func GetSwaggerDocument(w http.ResponseWriter, r *http.Request) {
	name := filepath.Join("./", "swagger.json")
	_, err := os.Stat(name)
	if err == nil || os.IsExist(err) {
		http.ServeFile(w, r, name)
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}
