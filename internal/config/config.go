package config

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strings"
)

const (
	EnvPrefix   = "MCBAPI_"
	EnvLocal    = "local"
	EnvDev      = "dev"
	EnvTest     = "test"
	EnvProd     = "prod"
	EnvDefault  = EnvLocal
	Environment = "ENVIRONMENT"
)

// var globalConfig map[string]any
var globalConfig *viper.Viper

// InitConfig initializes the configuration from default config folder and files
func InitConfig() error {
	return InitConfigWithFolder("", "")
}

// InitConfigWithFolder initializes the configuration usi
func InitConfigWithFolder(configfolder string, configfile string) error {
	v := viper.New()

	// default the environment so we know which .env.* file to pick up from non-prod environments
	v.SetDefault(Environment, EnvDefault)
	// see if there is an environment set in the OS env vars, either with or without a prefix
	envFromEnvironment := os.Getenv(EnvPrefix + Environment)
	if envFromEnvironment == "" {
		envFromEnvironment = os.Getenv(Environment)
	}
	if envFromEnvironment != "" {
		v.Set(Environment, strings.ToLower(envFromEnvironment))
	}
	fmt.Println("Environment:", v.GetString(Environment))

	// Set config name and paths for non-prod config (prod pulls from OS environment variables)
	if configfolder == "" {
		v.AddConfigPath("./config")
	} else {
		v.AddConfigPath(configfolder)
	}
	if configfile == "" {
		v.SetConfigName(v.GetString(Environment))
	} else {
		v.SetConfigName(configfile)
	}
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

func IsDevelopment() bool {
	return globalConfig.GetString(Environment) == EnvDev
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
