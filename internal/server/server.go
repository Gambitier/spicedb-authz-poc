package server

import (
	"context"
	"time"

	"github.com/Gambitier/spicedb-authz-poc/internal/config"
	"github.com/sirupsen/logrus"
)

// Server represents the main server that coordinates HTTP and gRPC servers
type Server struct {
	config *config.Config
	logger *logrus.Logger

	httpServer *HTTPServer
	grpcServer *GRPCServer
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config, logger *logrus.Logger) *Server {
	return &Server{
		config: cfg,
		logger: logger,
	}
}

// Start starts both HTTP and gRPC servers
func (s *Server) Start(ctx context.Context) error {
	// Initialize servers
	s.httpServer = NewHTTPServer(s.config, s.logger)
	s.grpcServer = NewGRPCServer(s.config, s.logger)

	// Register routes and services
	s.httpServer.RegisterRoutes()
	s.grpcServer.RegisterServices()

	// Start HTTP server
	go func() {
		if err := s.httpServer.Start(); err != nil {
			s.logger.Errorf("HTTP server error: %v", err)
		}
	}()

	// Start gRPC server
	go func() {
		if err := s.grpcServer.Start(); err != nil {
			s.logger.Errorf("gRPC server error: %v", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	return s.Shutdown(ctx)
}

// Shutdown gracefully shuts down both servers
func (s *Server) Shutdown(ctx context.Context) error {
	// Create a timeout context for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			s.logger.Errorf("HTTP server shutdown error: %v", err)
		}
	}

	// Shutdown gRPC server
	if s.grpcServer != nil {
		s.grpcServer.Shutdown()
	}

	return nil
}
