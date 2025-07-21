package config

import (
	"log/slog"
	"os"
	"strconv"
)

// Config holds all configuration for the MCP server
type Config struct {
	// Server configuration
	Host string
	Port int

	// Logging configuration
	LogLevel  string
	LogFormat string // "json" or "text"

	// Application configuration
	AppName    string
	AppVersion string

	// Kubernetes configuration
	KubeConfigPath string
	KubeContext    string
	K8sTimeout     string
}

// Load creates a new Config with values from environment variables
func Load() *Config {
	return &Config{
		Host:           getEnvString("MCP_HOST", "localhost"),
		Port:           getEnvInt("MCP_PORT", 8080),
		LogLevel:       getEnvString("LOG_LEVEL", "info"),
		LogFormat:      getEnvString("LOG_FORMAT", "json"),
		AppName:        getEnvString("APP_NAME", "idp-mcp-server"),
		AppVersion:     getEnvString("APP_VERSION", "dev"),
		KubeConfigPath: getEnvString("KUBE_CONFIG_PATH", ""),
		KubeContext:    getEnvString("KUBE_CONTEXT", ""),
		K8sTimeout:     getEnvString("K8S_TIMEOUT", "30s"),
	}
}

// Validate ensures the configuration is valid
func (c *Config) Validate() error {
	// Add validation logic here as needed
	return nil
}

// getEnvString returns the environment variable value or default
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt returns the environment variable value as int or default
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		slog.Warn("Invalid integer value for environment variable, using default",
			"key", key, "value", value, "default", defaultValue)
	}
	return defaultValue
}
