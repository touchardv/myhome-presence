package api

import (
	_ "embed"
	"net/http"
	"text/template"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/internal/config"
)

//go:embed  openapi.yaml.tmpl
var openapiYAML []byte

func getSwaggerDocument(w http.ResponseWriter, _ *http.Request, cfg config.Server) {
	t, err := template.New("openapi").Parse(string(openapiYAML))
	if err != nil {
		log.Fatal("Error parsing template: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data := make(map[string]interface{})
	data["address"] = getServerIPAddress(cfg.Address)
	data["port"] = cfg.Port
	t.Execute(w, data)
}
