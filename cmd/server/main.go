package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/Gambitier/spicedb-authz-poc/internal/config"
	"github.com/Gambitier/spicedb-authz-poc/internal/server"
	"github.com/Gambitier/spicedb-authz-poc/internal/services/authorization"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func init() {
	logger = logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
}

type CommandFlags struct {
	ConfigPath string
	Env        string
}

func NewCommandFlags() *CommandFlags {
	flags := &CommandFlags{}
	flag.StringVar(&flags.ConfigPath, "config", "default.yaml", "path to config file")
	flag.StringVar(&flags.Env, "env", "development", "environment")
	flag.Parse()
	return flags
}

func main() {
	flags := NewCommandFlags()

	// Load configuration
	cfg, err := config.LoadConfig(logger, flags.ConfigPath, flags.Env)
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	// Set log level
	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err != nil {
		logger.Fatalf("Failed to parse log level: %v", err)
	}
	logger.SetLevel(level)

	// Create server instance
	srv := server.NewServer(cfg, logger)

	testAuthorization(cfg)

	// Create context that listens for the interrupt signal from the OS
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server
	if err := srv.Start(ctx); err != nil {
		logger.Fatalf("Server error: %v", err)
	}
}

func testAuthorization(cfg *config.Config) {
	const schema = `
	definition user {}
	definition post {
		relation reader: user
		relation writer: user
		permission read = reader + writer
		permission write = writer
	}
`

	authorizationService, err := authorization.NewAuthorizationService(
		logger,
		cfg.SpiceDB,
		schema,
	)
	if err != nil {
		logger.Fatalf("Failed to create authorization service: %v", err)
	}

	err = authorizationService.WriteSchema()
	if err != nil {
		logger.Fatalf("Failed to write schema: %v", err)
	}

	err = authorizationService.WriteRelationships()
	if err != nil {
		logger.Fatalf("Failed to write relationships: %v", err)
	}

	err = authorizationService.CheckPermissions()
	if err != nil {
		logger.Fatalf("Failed to check permissions: %v", err)
	}
}
