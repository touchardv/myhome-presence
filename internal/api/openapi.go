package api

import (
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"text/template"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/internal/config"
)

//go:embed  openapi.yaml.tmpl
var openapiYAML []byte

func GetOpenAPISpecificationDocument(cfg config.Server) http.HandlerFunc {
	t, err := template.New("openapi").Parse(string(openapiYAML))
	if err != nil {
		log.Fatal("Error parsing template: ", err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		data := make(map[string]interface{})
		data["serverBaseURL"] = serverBaseURL(r, cfg)
		t.Execute(w, data)
	}
}

func GetSwaggerUIHandler(cfg config.Server, path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		scheme := ingressScheme(r, cfg)
		hostname := ingressHostname(r, cfg)
		port := ingressPort(r, cfg)
		url := fmt.Sprintf("%s/?url=%s://%s:%d%s", cfg.SwaggerUIURL, scheme, hostname, port, path)
		http.Redirect(w, r, url, http.StatusPermanentRedirect)
	}
}

func serverBaseURL(r *http.Request, cfg config.Server) string {
	scheme := ingressScheme(r, cfg)
	hostname := ingressHostname(r, cfg)
	port := ingressPort(r, cfg)
	return fmt.Sprintf("%s://%s:%d", scheme, hostname, port)
}

func ingressScheme(r *http.Request, cfg config.Server) string {
	protocol := r.Header.Get("X-Forwarded-Proto")
	if len(protocol) == 0 {
		// fall-back to configuration
		if cfg.SSL {
			return "https"
		} else {
			return "http"
		}
	}
	return protocol
}

func ingressHostname(r *http.Request, cfg config.Server) string {
	hostname := r.Header.Get("X-Forwarded-Host")
	if len(hostname) == 0 {
		// fall-back to configuration
		return hostnameOrIPAddress(cfg)
	}
	return hostname
}

func ingressPort(r *http.Request, cfg config.Server) uint {
	port := r.Header.Get("X-Forwarded-Port")
	if len(port) == 0 {
		// fall-back to configuration
		return cfg.Port
	}
	v, _ := strconv.Atoi(port)
	return uint(v)
}

func hostnameOrIPAddress(cfg config.Server) string {
	if len(cfg.Hostname) > 0 {
		return cfg.Hostname
	}
	if cfg.Address == "0.0.0.0" {
		addrs, err := net.InterfaceAddrs()
		if err == nil {
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						return ipnet.IP.String()
					}
				}
			}
		}
	} else {
		if len(cfg.Address) > 0 {
			return cfg.Address
		}
	}
	return "127.0.0.1"
}
