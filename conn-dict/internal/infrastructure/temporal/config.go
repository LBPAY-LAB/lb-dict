package temporal

import (
	"fmt"
	"os"
)

// Config holds the configuration for Temporal connection
type Config struct {
	// Host is the Temporal server host
	Host string `mapstructure:"host"`
	// Port is the Temporal server port
	Port int `mapstructure:"port"`
	// Namespace is the Temporal namespace to use
	Namespace string `mapstructure:"namespace"`
	// TaskQueue is the default task queue name
	TaskQueue string `mapstructure:"task_queue"`
}

// DefaultConfig returns a default Temporal configuration
func DefaultConfig() Config {
	return Config{
		Host:      getEnvOrDefault("TEMPORAL_HOST", "localhost"),
		Port:      getEnvIntOrDefault("TEMPORAL_PORT", 7233),
		Namespace: getEnvOrDefault("TEMPORAL_NAMESPACE", "default"),
		TaskQueue: getEnvOrDefault("TEMPORAL_TASK_QUEUE", "conn-dict-task-queue"),
	}
}

// HostPort returns the host:port string
func (c Config) HostPort() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Validate validates the configuration
func (c Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("temporal host is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("temporal port must be between 1 and 65535")
	}
	if c.Namespace == "" {
		return fmt.Errorf("temporal namespace is required")
	}
	if c.TaskQueue == "" {
		return fmt.Errorf("temporal task queue is required")
	}
	return nil
}

// getEnvOrDefault returns the environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault returns the environment variable value as int or default
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}
