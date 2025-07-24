package config

import (
	"log/slog"
	"os"
	"strconv"
)

const (
	// AppName is the application name
	AppName = "idp-mcp-server"
)

var (
	// Version is set at build time using -ldflags "-X github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/config.Version=v1.0.0"
	Version = "dev"
)

// Config holds all configuration for the MCP server
type Config struct {
	// Server configuration
	Host string
	Port int

	// Logging configuration
	LogLevel  string
	LogFormat string // "json" or "text"

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
		KubeConfigPath: getEnvString("KUBECONFIG", "~/.kube/config"),
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
