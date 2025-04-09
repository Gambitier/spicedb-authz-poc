package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/Gambitier/spicedb-authz-poc/internal/config"
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

	// Run migrations
	if err := runMigrations(cfg); err != nil {
		logger.Fatalf("Failed to run migrations: %v", err)
	}

	logger.Info("Migrations completed successfully")
}

func runMigrations(cfg *config.Config) error {
	if cfg.SpiceDB.Datastore == nil {
		return fmt.Errorf("datastore configuration is required")
	}

	logger.Infof("Running migrations for datastore: %s", cfg.SpiceDB.Datastore.Engine)
	logger.Infof("Using connection URI: %s", cfg.SpiceDB.Datastore.URI)

	// Execute spicedb migrate command
	args := []string{
		"migrate", "head",
		"--datastore-engine", string(cfg.SpiceDB.Datastore.Engine),
		"--datastore-conn-uri", cfg.SpiceDB.Datastore.URI,
	}

	cmd := exec.Command("spicedb", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}
