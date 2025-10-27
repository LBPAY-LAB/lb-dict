package di

import (
	"fmt"
	"os"
	"time"

	"github.com/lbpay-lab/conn-bridge/internal/api/grpc"
	"github.com/lbpay-lab/conn-bridge/internal/application/usecases"
	"github.com/lbpay-lab/conn-bridge/internal/domain/interfaces"
	"github.com/lbpay-lab/conn-bridge/internal/infrastructure/bacen"
	"github.com/lbpay-lab/conn-bridge/internal/infrastructure/circuitbreaker"
	"github.com/lbpay-lab/conn-bridge/internal/infrastructure/pulsar"
	"github.com/sirupsen/logrus"
)

// Container holds all dependencies for the application
type Container struct {
	// Infrastructure
	BacenClient      interfaces.BacenClient
	MessagePublisher interfaces.MessagePublisher
	CircuitBreaker   interfaces.CircuitBreaker

	// Use Cases
	CreateEntryUseCase *usecases.CreateEntryUseCase
	QueryEntryUseCase  *usecases.QueryEntryUseCase
	DeleteEntryUseCase *usecases.DeleteEntryUseCase
	CreateClaimUseCase *usecases.CreateClaimUseCase

	// API
	GRPCServer *grpc.Server
}

// Config holds the configuration for the container
type Config struct {
	// Bacen configuration
	BacenBaseURL  string
	BacenTimeout  time.Duration
	BacenAPIKey   string
	BacenCertPath string
	BacenKeyPath  string

	// Pulsar configuration
	PulsarBrokerURL string
	PulsarTimeout   time.Duration

	// Circuit Breaker configuration
	CircuitBreakerName        string
	CircuitBreakerMaxFailures uint32 // Max consecutive failures before opening (default: 5)
	CircuitBreakerTimeout     time.Duration // Timeout before half-open (default: 60s)
	CircuitBreakerMaxRequests uint32 // Max requests in half-open state (default: 3)

	// gRPC configuration
	GRPCPort int
}

// NewContainer creates a new dependency injection container
func NewContainer(config *Config) (*Container, error) {
	container := &Container{}

	// Initialize infrastructure layer
	if err := container.initInfrastructure(config); err != nil {
		return nil, fmt.Errorf("failed to initialize infrastructure: %w", err)
	}

	// Initialize application layer
	if err := container.initApplication(); err != nil {
		return nil, fmt.Errorf("failed to initialize application: %w", err)
	}

	// Initialize API layer
	if err := container.initAPI(config); err != nil {
		return nil, fmt.Errorf("failed to initialize API: %w", err)
	}

	return container, nil
}

// initInfrastructure initializes the infrastructure layer
func (c *Container) initInfrastructure(config *Config) error {
	// Initialize logger for infrastructure components
	logger := &logrus.Logger{
		Out:       os.Stdout,
		Formatter: new(logrus.JSONFormatter),
		Level:     logrus.InfoLevel,
	}

	// Initialize Bacen HTTP client (base client without circuit breaker)
	baseBacenClient, err := bacen.NewHTTPClient(&bacen.Config{
		BaseURL:  config.BacenBaseURL,
		Timeout:  config.BacenTimeout,
		APIKey:   config.BacenAPIKey,
		CertPath: config.BacenCertPath,
		KeyPath:  config.BacenKeyPath,
	})
	if err != nil {
		return fmt.Errorf("failed to create Bacen client: %w", err)
	}

	// Initialize Circuit Breaker with proper configuration
	c.CircuitBreaker = circuitbreaker.NewGoBreakerAdapter(&circuitbreaker.Config{
		Name:        config.CircuitBreakerName,
		MaxRequests: config.CircuitBreakerMaxRequests, // Half-open max requests
		Timeout:     config.CircuitBreakerTimeout,     // Before attempting half-open
		Interval:    60 * time.Second,                 // Clear counts interval
		MaxFailures: config.CircuitBreakerMaxFailures, // Consecutive errors before opening
		Logger:      logger,
	})

	// Wrap Bacen client with circuit breaker protection
	c.BacenClient = bacen.NewCircuitBreakerClient(baseBacenClient, c.CircuitBreaker, logger)

	// Initialize Pulsar publisher
	pulsarPublisher, err := pulsar.NewPublisher(&pulsar.Config{
		BrokerURL:               config.PulsarBrokerURL,
		OperationTimeout:        config.PulsarTimeout,
		ConnectionTimeout:       10 * time.Second,
		MaxConnectionsPerBroker: 1,
	})
	if err != nil {
		return fmt.Errorf("failed to create Pulsar publisher: %w", err)
	}
	c.MessagePublisher = pulsarPublisher

	return nil
}

// initApplication initializes the application layer (use cases)
func (c *Container) initApplication() error {
	// Initialize use cases
	c.CreateEntryUseCase = usecases.NewCreateEntryUseCase(
		c.BacenClient,
		c.MessagePublisher,
		c.CircuitBreaker,
	)

	c.QueryEntryUseCase = usecases.NewQueryEntryUseCase(
		c.BacenClient,
		c.CircuitBreaker,
	)

	c.DeleteEntryUseCase = usecases.NewDeleteEntryUseCase(
		c.BacenClient,
		c.MessagePublisher,
		c.CircuitBreaker,
	)

	c.CreateClaimUseCase = usecases.NewCreateClaimUseCase(
		c.BacenClient,
		c.MessagePublisher,
		c.CircuitBreaker,
	)

	return nil
}

// initAPI initializes the API layer (gRPC server)
func (c *Container) initAPI(config *Config) error {
	grpcServer, err := grpc.NewServer(&grpc.Config{
		Port:               config.GRPCPort,
		CreateEntryUseCase: c.CreateEntryUseCase,
		QueryEntryUseCase:  c.QueryEntryUseCase,
		DeleteEntryUseCase: c.DeleteEntryUseCase,
		CreateClaimUseCase: c.CreateClaimUseCase,
	})
	if err != nil {
		return fmt.Errorf("failed to create gRPC server: %w", err)
	}
	c.GRPCServer = grpcServer

	return nil
}

// Close closes all resources in the container
func (c *Container) Close() error {
	if c.MessagePublisher != nil {
		if err := c.MessagePublisher.Close(); err != nil {
			return fmt.Errorf("failed to close message publisher: %w", err)
		}
	}

	if c.GRPCServer != nil {
		c.GRPCServer.Stop()
	}

	return nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		BacenBaseURL:              "https://api.bacen.gov.br/dict",
		BacenTimeout:              30 * time.Second,
		PulsarBrokerURL:           "pulsar://localhost:6650",
		PulsarTimeout:             30 * time.Second,
		CircuitBreakerName:        "bacen-circuit-breaker",
		CircuitBreakerMaxFailures: 5,                // 5 consecutive failures
		CircuitBreakerTimeout:     60 * time.Second, // 60s before half-open
		CircuitBreakerMaxRequests: 3,                // 3 requests in half-open
		GRPCPort:                  50051,
	}
}
