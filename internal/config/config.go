package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration values
type Config struct {
	Environment string
	GinMode     string
	Server      ServerConfig
	Database    DatabaseConfig
	Logging     LoggingConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Name     string
	IP       string
	Port     string
	Hostname string
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	URL      string
	User     string
	Password string
}

// LoggingConfig holds logging-related configuration
type LoggingConfig struct {
	Level    string
	Format   string
	Output   string
	FilePath string
}

var globalConfig *Config

// InitConfig initializes the configuration using Viper
func InitConfig() error {
	v := viper.New()

	// Set config name and paths
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	// Set environment variable prefix and enable automatic env reading
	v.SetEnvPrefix("MCBAPI_")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set default values
	v.SetDefault("environment", "dev")
	v.SetDefault("gin_mode", "debug")
	v.SetDefault("server.name", "unknown")
	v.SetDefault("server.ip", "unknown")
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.hostname", "http://localhost:8080/api")
	v.SetDefault("database.url", "postgres://localhost:5432/millcheckdb")
	v.SetDefault("database.user", "mcuser")
	v.SetDefault("database.password", "mcuser")
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")
	v.SetDefault("logging.file_path", "/var/log/mcb-api.log")

	// Read config file (optional - will use defaults and env vars if not found)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is okay, we'll use defaults and env vars
	}

	// Unmarshal into config struct
	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Special handling for legacy environment variables
	if ginMode := v.GetString("GIN_MODE"); ginMode != "" {
		config.GinMode = ginMode
	}
	if serverName := v.GetString("SERVER_NAME"); serverName != "" {
		config.Server.Name = serverName
	}
	if serverIP := v.GetString("SERVER_IP"); serverIP != "" {
		config.Server.IP = serverIP
	}
	if logLevel := v.GetString("LOG_LEVEL"); logLevel != "" {
		config.Logging.Level = logLevel
	}
	if logFormat := v.GetString("LOG_FORMAT"); logFormat != "" {
		config.Logging.Format = logFormat
	}
	if logOutput := v.GetString("LOG_OUTPUT"); logOutput != "" {
		config.Logging.Output = logOutput
	}
	if logFilePath := v.GetString("LOG_FILE_PATH"); logFilePath != "" {
		config.Logging.FilePath = logFilePath
	}

	globalConfig = config
	return nil
}

// GetConfig returns the global configuration
func GetConfig() *Config {
	if globalConfig == nil {
		// Initialize with defaults if not already initialized
		if err := InitConfig(); err != nil {
			panic(fmt.Sprintf("Failed to initialize config: %v", err))
		}
	}
	return globalConfig
}

// GetString returns a string configuration value
func GetString(key string) string {
	config := GetConfig()
	switch key {
	case "GIN_MODE":
		return config.GinMode
	case "SERVER_NAME":
		return config.Server.Name
	case "SERVER_IP":
		return config.Server.IP
	case "LOG_LEVEL":
		return config.Logging.Level
	case "LOG_FORMAT":
		return config.Logging.Format
	case "LOG_OUTPUT":
		return config.Logging.Output
	case "LOG_FILE_PATH":
		return config.Logging.FilePath
	default:
		// Fallback to direct viper access for other keys
		return viper.GetString(key)
	}
}

// GetStringWithDefault returns a string configuration value with a default
func GetStringWithDefault(key, defaultValue string) string {
	value := GetString(key)
	if value == "" {
		return defaultValue
	}
	return value
}
