package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/touchardv/myhome-presence/internal/config"
)

func TestGetOpenAPISpecificationDocument(t *testing.T) {
	req := httptest.NewRequest("GET", "http://foo", nil)
	req.Header = map[string][]string{
		"X-Forwarded-Proto": {"https"},
		"X-Forwarded-Port":  {"443"},
	}
	rr := httptest.NewRecorder()

	cfg := config.Server{
		Address:      "0.0.0.0",
		Hostname:     "api.server.com",
		Port:         8080,
		SwaggerUIURL: "https://swagger-ui",
	}
	h := GetOpenAPISpecificationDocument(cfg)
	h(rr, req)

	res := rr.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	b, err := io.ReadAll(rr.Result().Body)
	assert.NoError(t, err)
	body := string(b)
	assert.Contains(t, body, "https://api.server.com:443/api")
}

func TestGetSwaggerUIHandlerWithDirectAccess(t *testing.T) {
	req := httptest.NewRequest("GET", "http://foo", nil)
	rr := httptest.NewRecorder()

	cfg := config.Server{
		Port:         8080,
		SwaggerUIURL: "https://swagger-ui",
	}
	h := GetSwaggerUIHandler(cfg, "/api/docs")
	h(rr, req)

	res := rr.Result()
	assert.Equal(t, http.StatusPermanentRedirect, res.StatusCode)
	location := res.Header.Get("Location")
	assert.Equal(t, "https://swagger-ui/?url=http://127.0.0.1:8080/api/docs", location)
}
