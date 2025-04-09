package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Gambitier/spicedb-authz-poc/internal/config"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// HTTPServer represents the HTTP server
type HTTPServer struct {
	config *config.Config
	logger *logrus.Logger
	server *http.Server
	router *mux.Router
}

// NewHTTPServer creates a new HTTP server instance
func NewHTTPServer(cfg *config.Config, logger *logrus.Logger) *HTTPServer {
	router := mux.NewRouter()

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.API.HTTPPort),
		Handler:      router,
		ReadTimeout:  cfg.API.ReadTimeout,
		WriteTimeout: cfg.API.WriteTimeout,
		IdleTimeout:  cfg.API.IdleTimeout,
	}

	return &HTTPServer{
		config: cfg,
		logger: logger,
		server: server,
		router: router,
	}
}

// Start starts the HTTP server
func (s *HTTPServer) Start() error {
	s.logger.Infof("Starting HTTP server on port %d", s.config.API.HTTPPort)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server...")
	return s.server.Shutdown(ctx)
}

// Router returns the mux router for adding routes
func (s *HTTPServer) Router() *mux.Router {
	return s.router
}

// RegisterRoutes registers all HTTP routes
func (s *HTTPServer) RegisterRoutes() {
	// Add your HTTP routes here
	// Example:
	// s.router.HandleFunc("/v1/check-permission", s.handleCheckPermission).Methods(http.MethodPost)
	// s.router.HandleFunc("/v1/set-manager", s.handleSetManager).Methods(http.MethodPost)

	// health check
	s.router.HandleFunc("/health", s.handleHealthCheck).Methods(http.MethodGet)
}

func (s *HTTPServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
