package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/device"
	"github.com/touchardv/myhome-presence/docs"
)

type apiContext struct {
	registry *device.Registry
}

// Server is a wrapper around the router and the HTTP server.
type Server struct {
	apiContext
	server  *http.Server
	router  *mux.Router
	stopped chan bool
}

// NewServer creates and initializes a new API server.
func NewServer(r *device.Registry) *Server {
	apiContext := apiContext{r}
	router := mux.NewRouter()

	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"content-type"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowCredentials())

	router.Use(cors)
	router.HandleFunc("/health-check", healthCheck).Methods("GET")
	router.HandleFunc("/api/docs", docs.GetSwaggerDocument).Methods("GET")
	router.HandleFunc("/api/devices", apiContext.listDevices).Methods("GET")

	server := &http.Server{
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      router,
	}
	return &Server{
		server:  server,
		router:  router,
		stopped: make(chan bool, 1),
	}
}

// Start runs the HTTP server (in the background).
func (s *Server) Start() {
	log.Info("Starting: http server")
	go func() {
		server := s.server
		log.Info("Listening on: " + server.Addr)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Error(err)
		}
		s.stopped <- true
	}()
}

// Stop shutdowns the HTTP server.
func (s *Server) Stop() {
	log.Info("Stopping: http server")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	s.server.Shutdown(ctx)
	<-s.stopped
	log.Info("Stopped: http server")
}
