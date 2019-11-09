package api

import (
	"net/http"
)

// healthCheck handles a request by returning an empty response with status "No Content".
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
