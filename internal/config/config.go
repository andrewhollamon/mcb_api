package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strings"
)

const (
	EnvPrefix  = "MCBAPI_"
	EnvLocal   = "local"
	EnvDev     = "dev"
	EnvTest    = "test"
	EnvProd    = "prod"
	EnvDefault = EnvLocal
)

// var globalConfig map[string]any
var globalConfig *viper.Viper

// InitConfig initializes the configuration using Viper
func InitConfig() error {
	v := viper.New()

	// default the environment so we know which .env.* file to pick up from non-prod environments
	v.SetDefault("ENVIRONMENT", EnvDefault)
	// see if there is an environment set in the OS env vars, either with or without a prefix
	envFromEnvironment := os.Getenv(EnvPrefix + "ENVIRONMENT")
	if envFromEnvironment == "" {
		envFromEnvironment = os.Getenv("ENVIRONMENT")
	}
	if envFromEnvironment != "" {
		v.Set("ENVIRONMENT", strings.ToLower(envFromEnvironment))
	}
	fmt.Println("Environment:", v.GetString("ENVIRONMENT"))

	// Set config name and paths for non-prod config (prod pulls from OS environment variables)
	v.AddConfigPath("./config")
	v.SetConfigName(v.GetString("ENVIRONMENT"))
	v.SetConfigType("env")

	// Set environment variable prefix and enable automatic env reading
	v.SetEnvPrefix(EnvPrefix)
	v.AutomaticEnv()
	//v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file (optional - will use defaults and env vars if not found)
	if err := v.ReadInConfig(); err != nil {
		fmt.Println("Error reading config file:", err)
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is okay, we'll use defaults and env vars
	}
	fmt.Println("Using config file:", v.ConfigFileUsed())

	globalConfig = v
	return nil
}

// GetConfig returns the global configuration
func GetConfig() *viper.Viper {
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
	return globalConfig.GetString(key)
}

// GetStringWithDefault returns a string configuration value with a default
func GetStringWithDefault(key, defaultValue string) string {
	value := GetString(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// DumpConfig prints the entire processed configuration
func DumpConfig() {
	fmt.Println("=== Configuration Dump ===")

	// Pretty print the config struct
	for k, v := range globalConfig.AllSettings() {
		fmt.Printf("%s: %v\n", k, v)
	}
	fmt.Println("=========================")
}
