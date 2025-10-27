package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/lbpay-lab/conn-bridge/internal/di"
)

var (
	log = logrus.New()
)

func main() {
	// Initialize logger
	initLogger()

	log.Info("Starting conn-bridge service...")

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize dependency injection container
	log.Info("Initializing dependencies...")
	container, err := di.NewContainer(config)
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}
	defer func() {
		if err := container.Close(); err != nil {
			log.Errorf("Error closing container: %v", err)
		}
	}()

	// Start gRPC server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		log.Infof("Starting gRPC server on port %d...", config.GRPCPort)
		if err := container.GRPCServer.Start(); err != nil {
			errChan <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	// Wait for interrupt signal or server error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		log.Errorf("Server error: %v", err)
	case sig := <-sigChan:
		log.Infof("Received signal: %v", sig)
	}

	// Graceful shutdown
	log.Info("Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Close container (which stops gRPC server)
	if err := container.Close(); err != nil {
		log.Errorf("Error during shutdown: %v", err)
	}

	// Wait for context timeout or completion
	<-ctx.Done()
	log.Info("Server stopped")
}

// loadConfig loads configuration from environment and config files
func loadConfig() (*di.Config, error) {
	// Set config file paths
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/conn-bridge")

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		log.Warn("No config file found, using environment variables and defaults")
	}

	// Enable environment variable override
	viper.AutomaticEnv()
	viper.SetEnvPrefix("CONN_BRIDGE")

	// Build configuration
	config := &di.Config{
		// Bacen configuration
		BacenBaseURL:  getEnvOrDefault("BACEN_BASE_URL", "https://api-dict.bcb.gov.br"),
		BacenTimeout:  getDurationOrDefault("BACEN_TIMEOUT", 30*time.Second),
		BacenAPIKey:   viper.GetString("BACEN_API_KEY"),
		BacenCertPath: viper.GetString("BACEN_CERT_PATH"),
		BacenKeyPath:  viper.GetString("BACEN_KEY_PATH"),

		// Pulsar configuration
		PulsarBrokerURL: getEnvOrDefault("PULSAR_BROKER_URL", "pulsar://localhost:6650"),
		PulsarTimeout:   getDurationOrDefault("PULSAR_TIMEOUT", 30*time.Second),

		// Circuit Breaker configuration
		CircuitBreakerName:        getEnvOrDefault("CIRCUIT_BREAKER_NAME", "bacen-circuit-breaker"),
		CircuitBreakerMaxFailures: getUint32OrDefault("CIRCUIT_BREAKER_MAX_FAILURES", 5),
		CircuitBreakerTimeout:     getDurationOrDefault("CIRCUIT_BREAKER_TIMEOUT", 60*time.Second),
		CircuitBreakerMaxRequests: getUint32OrDefault("CIRCUIT_BREAKER_MAX_REQUESTS", 3),

		// gRPC configuration
		GRPCPort: getIntOrDefault("GRPC_PORT", 50051),
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	log.WithFields(logrus.Fields{
		"bacen_url":         config.BacenBaseURL,
		"pulsar_url":        config.PulsarBrokerURL,
		"grpc_port":         config.GRPCPort,
		"cb_name":           config.CircuitBreakerName,
		"cb_max_failures":   config.CircuitBreakerMaxFailures,
		"cb_timeout":        config.CircuitBreakerTimeout,
		"cb_max_requests":   config.CircuitBreakerMaxRequests,
	}).Info("Configuration loaded")

	return config, nil
}

// validateConfig validates the configuration
func validateConfig(config *di.Config) error {
	if config.BacenBaseURL == "" {
		return fmt.Errorf("Bacen base URL is required")
	}
	if config.PulsarBrokerURL == "" {
		return fmt.Errorf("Pulsar broker URL is required")
	}
	if config.GRPCPort <= 0 || config.GRPCPort > 65535 {
		return fmt.Errorf("invalid gRPC port: %d", config.GRPCPort)
	}
	return nil
}

// initLogger initializes the logger
func initLogger() {
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)

	// Set log level from environment
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := viper.GetString(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntOrDefault(key string, defaultValue int) int {
	if value := viper.GetInt(key); value != 0 {
		return value
	}
	return defaultValue
}

func getDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := viper.GetDuration(key); value != 0 {
		return value
	}
	return defaultValue
}

func getUint32OrDefault(key string, defaultValue uint32) uint32 {
	if value := viper.GetUint32(key); value != 0 {
		return value
	}
	return defaultValue
}
