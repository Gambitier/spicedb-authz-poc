package server

import (
	"fmt"
	"net"

	"github.com/Gambitier/spicedb-authz-poc/internal/config"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GRPCServer represents the gRPC server
type GRPCServer struct {
	config *config.Config
	logger *logrus.Logger
	server *grpc.Server
}

// NewGRPCServer creates a new gRPC server instance
func NewGRPCServer(cfg *config.Config, logger *logrus.Logger) *GRPCServer {
	// Create gRPC server with default options
	server := grpc.NewServer()

	return &GRPCServer{
		config: cfg,
		logger: logger,
		server: server,
	}
}

// Start starts the gRPC server
func (s *GRPCServer) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.API.GRPCPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Register reflection service for development
	reflection.Register(s.server)

	s.logger.Infof("Starting gRPC server on port %d", s.config.API.GRPCPort)
	return s.server.Serve(lis)
}

// Shutdown gracefully shuts down the gRPC server
func (s *GRPCServer) Shutdown() {
	s.logger.Info("Shutting down gRPC server...")
	s.server.GracefulStop()
}

// Server returns the gRPC server instance for registering services
func (s *GRPCServer) Server() *grpc.Server {
	return s.server
}

// RegisterServices registers all gRPC services
func (s *GRPCServer) RegisterServices() {
	// Add your gRPC services here
	// Example:
	// pb.RegisterAuthorizationServer(s.server, s.authService)
}
