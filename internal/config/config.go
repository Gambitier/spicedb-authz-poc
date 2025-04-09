package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

// Config represents the main configuration structure
type Config struct {
	API     *APIConfig     `mapstructure:"api" validate:"required"`
	SpiceDB *SpiceDBConfig `mapstructure:"spicedb" validate:"required"`
	Metrics *MetricsConfig `mapstructure:"metrics" validate:"required"`
	Logging *LoggingConfig `mapstructure:"logging" validate:"required"`
}

// APIConfig represents the API server configuration
type APIConfig struct {
	HTTPPort     uint16        `mapstructure:"http_port" validate:"required,min=1,max=65535"`
	GRPCPort     uint16        `mapstructure:"grpc_port" validate:"required,min=1,max=65535"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" validate:"required"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" validate:"required"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout" validate:"required"`
}

type DatastoreEngine string

const (
	Postgres DatastoreEngine = "postgresql"
)

type DatastoreConfig struct {
	Engine          DatastoreEngine `mapstructure:"engine" validate:"required,oneof=postgresql"`
	URI             string          `mapstructure:"uri" validate:"required"`
	MaxOpenConns    int             `mapstructure:"max_open_conns" validate:"required,min=1"`
	MaxIdleConns    int             `mapstructure:"max_idle_conns" validate:"required,min=1"`
	ConnMaxLifetime time.Duration   `mapstructure:"conn_max_lifetime" validate:"required"`
	GCWindow        time.Duration   `mapstructure:"gc_window" validate:"required"`
}

// SpiceDBConfig represents the SpiceDB connection configuration
type SpiceDBConfig struct {
	Host           string           `mapstructure:"host" validate:"required"`
	Port           uint16           `mapstructure:"port" validate:"required,min=1,max=65535"`
	PresharedKey   string           `mapstructure:"preshared_key" validate:"required"`
	MaxRetries     int              `mapstructure:"max_retries" validate:"required,min=1"`
	RetryDelay     time.Duration    `mapstructure:"retry_delay" validate:"required"`
	RequestTimeout time.Duration    `mapstructure:"request_timeout" validate:"required"`
	Datastore      *DatastoreConfig `mapstructure:"datastore" validate:"required"`
}

// MetricsConfig represents the metrics configuration
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    uint16 `mapstructure:"port" validate:"required_if=Enabled true,min=1,max=65535"`
	Path    string `mapstructure:"path" validate:"required_if=Enabled true"`
}

// LoggingConfig represents the logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level" validate:"required,oneof=debug info warn error"`
	Format string `mapstructure:"format" validate:"required,oneof=json text"`
}

// LoadConfig loads and validates the configuration
func LoadConfig(logger *logrus.Logger, relConfigPath string, env string) (*Config, error) {
	// Get the absolute path to the config file
	configPath, err := filepath.Abs(relConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for config file: %w", err)
	}

	// Get the directory of the config file
	configDir := filepath.Dir(configPath)
	configName := filepath.Base(configPath)
	configExt := filepath.Ext(configPath)
	configNameWithoutExt := configName[:len(configName)-len(configExt)]

	// Initialize viper
	v := viper.New()
	v.SetConfigName(configNameWithoutExt)
	v.SetConfigType("yaml")
	v.AddConfigPath(configDir)

	// Set defaults
	setDefaults(v)

	// Try to read environment-specific config
	envConfigPath := filepath.Join(configDir, fmt.Sprintf("%s.%s%s", configNameWithoutExt, env, configExt))
	if _, err := os.Stat(envConfigPath); err == nil {
		v.SetConfigName(fmt.Sprintf("%s.%s", configNameWithoutExt, env))
		if err := v.MergeInConfig(); err != nil {
			return nil, fmt.Errorf("failed to merge environment config: %w", err)
		}
	}

	logger.Infof("Using config file: %s", envConfigPath)

	// Read the default config
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal the config
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate the config
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults sets default values for configuration
func setDefaults(v *viper.Viper) {
	// API defaults
	v.SetDefault("api.http_port", 8085)
	v.SetDefault("api.grpc_port", 8086)
	v.SetDefault("api.read_timeout", "5s")
	v.SetDefault("api.write_timeout", "5s")
	v.SetDefault("api.idle_timeout", "120s")

	// SpiceDB defaults
	v.SetDefault("spicedb.host", "localhost")
	v.SetDefault("spicedb.port", 50051)
	v.SetDefault("spicedb.max_retries", 3)
	v.SetDefault("spicedb.retry_delay", "1s")
	v.SetDefault("spicedb.request_timeout", "5s")
	// SpiceDB datastore defaults
	v.SetDefault("spicedb.datastore.engine", "postgresql")
	v.SetDefault("spicedb.datastore.max_open_conns", 20)
	v.SetDefault("spicedb.datastore.max_idle_conns", 10)
	v.SetDefault("spicedb.datastore.conn_max_lifetime", "30m")
	v.SetDefault("spicedb.datastore.gc_window", "24h")

	// Metrics defaults
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.port", 9093)
	v.SetDefault("metrics.path", "/metrics")

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
}

// validateConfig validates the configuration using struct tags
func validateConfig(cfg *Config) error {
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrors {
				return fmt.Errorf("validation failed for field %s: %s", e.Field(), e.Tag())
			}
		}
		return fmt.Errorf("validation failed: %w", err)
	}
	return nil
}
