package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/touchardv/myhome-presence/config"
	"github.com/touchardv/myhome-presence/device"
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
func NewServer(cfg config.Server, r *device.Registry) *Server {
	apiContext := apiContext{r}
	router := mux.NewRouter()

	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"content-type"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"DELETE", "GET", "POST", "PUT"}),
		handlers.AllowCredentials())

	url := fmt.Sprintf("%s/?url=http://%s:%d/api/docs", cfg.SwaggerUIURL, getServerIPAddress(cfg.Address), cfg.Port)
	router.Handle("/", http.RedirectHandler(url, http.StatusPermanentRedirect)).Methods("GET")
	router.HandleFunc("/health-check", healthCheck).Methods("GET")
	router.HandleFunc("/api/docs", func(w http.ResponseWriter, r *http.Request) {
		getSwaggerDocument(w, r, cfg)
	}).Methods("GET")
	router.HandleFunc("/api/devices", apiContext.registerDevice).Methods("POST")
	router.HandleFunc("/api/devices/{id}", apiContext.unregisterDevice).Methods("DELETE")
	router.HandleFunc("/api/devices/{id}", apiContext.findDevice).Methods("GET")
	router.HandleFunc("/api/devices/{id}", apiContext.executeDeviceAction).Methods("POST")
	router.HandleFunc("/api/devices/{id}", apiContext.updateDevice).Methods("PUT")
	router.HandleFunc("/api/devices", apiContext.queryDevices).Methods("GET")

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Address, cfg.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      cors(router),
	}
	return &Server{
		server:  server,
		router:  router,
		stopped: make(chan bool, 1),
	}
}

func getServerIPAddress(cfgAddr string) string {
	if cfgAddr == "0.0.0.0" {
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
		return cfgAddr
	}
	return "127.0.0.1"
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
