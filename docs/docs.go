package docs

import (
	_ "embed"
	"net/http"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"
)

//go:embed  openapi.yaml.tmpl
var openapiYAML []byte

// GetSwaggerDocument servers the Swagger JSON file.
func GetSwaggerDocument(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("openapi").Parse(string(openapiYAML))
	if err != nil {
		log.Fatal("Error parsing template: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data := make(map[string]interface{})
	data["date"] = time.Now().UTC().Format(time.RFC3339)
	t.Execute(w, data)
}
