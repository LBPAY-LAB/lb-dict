package temporal

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/client"
)

// ClientConfig holds the configuration for Temporal client
type ClientConfig struct {
	HostPort  string
	Namespace string
	Logger    *logrus.Logger
}

// Client wraps the Temporal client with additional functionality
type Client struct {
	client    client.Client
	namespace string
	logger    *logrus.Logger
}

// NewClient creates a new Temporal client with the given configuration
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.HostPort == "" {
		cfg.HostPort = "localhost:7233"
	}

	if cfg.Namespace == "" {
		cfg.Namespace = "default"
	}

	if cfg.Logger == nil {
		cfg.Logger = logrus.New()
		cfg.Logger.SetLevel(logrus.InfoLevel)
	}

	cfg.Logger.WithFields(logrus.Fields{
		"host":      cfg.HostPort,
		"namespace": cfg.Namespace,
	}).Info("Connecting to Temporal server")

	// Create Temporal client options
	clientOptions := client.Options{
		HostPort:  cfg.HostPort,
		Namespace: cfg.Namespace,
		Logger:    NewTemporalLogger(cfg.Logger),
	}

	// Create the client
	c, err := client.Dial(clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %w", err)
	}

	cfg.Logger.Info("Successfully connected to Temporal server")

	return &Client{
		client:    c,
		namespace: cfg.Namespace,
		logger:    cfg.Logger,
	}, nil
}

// GetClient returns the underlying Temporal client
func (c *Client) GetClient() client.Client {
	return c.client
}

// GetNamespace returns the configured namespace
func (c *Client) GetNamespace() string {
	return c.namespace
}

// Close closes the Temporal client connection
func (c *Client) Close() {
	if c.client != nil {
		c.logger.Info("Closing Temporal client connection")
		c.client.Close()
	}
}

// HealthCheck performs a health check on the Temporal connection
func (c *Client) HealthCheck(ctx context.Context) error {
	if c.client == nil {
		return fmt.Errorf("client is not initialized")
	}

	// Try to check server capabilities - this will fail if connection is down
	c.logger.Debug("Performing Temporal health check")

	// Use CheckHealth method if available, otherwise connection itself validates health
	_, err := c.client.CheckHealth(ctx, &client.CheckHealthRequest{})
	if err != nil {
		c.logger.WithError(err).Error("Temporal health check failed")
		return fmt.Errorf("temporal health check failed: %w", err)
	}

	c.logger.Debug("Temporal health check passed")
	return nil
}

// ExecuteWorkflow executes a workflow with the given options
func (c *Client) ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (client.WorkflowRun, error) {
	c.logger.WithFields(logrus.Fields{
		"workflow_id":   options.ID,
		"task_queue":    options.TaskQueue,
		"workflow_type": fmt.Sprintf("%T", workflow),
	}).Info("Executing workflow")

	run, err := c.client.ExecuteWorkflow(ctx, options, workflow, args...)
	if err != nil {
		c.logger.WithError(err).Error("Failed to execute workflow")
		return nil, fmt.Errorf("failed to execute workflow: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"workflow_id": options.ID,
		"run_id":      run.GetRunID(),
	}).Info("Workflow started successfully")

	return run, nil
}

// GetWorkflow gets a handle to an existing workflow
func (c *Client) GetWorkflow(ctx context.Context, workflowID string, runID string) client.WorkflowRun {
	return c.client.GetWorkflow(ctx, workflowID, runID)
}

// CancelWorkflow cancels a running workflow
func (c *Client) CancelWorkflow(ctx context.Context, workflowID string, runID string) error {
	c.logger.WithFields(logrus.Fields{
		"workflow_id": workflowID,
		"run_id":      runID,
	}).Info("Cancelling workflow")

	err := c.client.CancelWorkflow(ctx, workflowID, runID)
	if err != nil {
		c.logger.WithError(err).Error("Failed to cancel workflow")
		return fmt.Errorf("failed to cancel workflow: %w", err)
	}

	c.logger.Info("Workflow cancelled successfully")
	return nil
}

// TerminateWorkflow terminates a running workflow
func (c *Client) TerminateWorkflow(ctx context.Context, workflowID string, runID string, reason string) error {
	c.logger.WithFields(logrus.Fields{
		"workflow_id": workflowID,
		"run_id":      runID,
		"reason":      reason,
	}).Info("Terminating workflow")

	err := c.client.TerminateWorkflow(ctx, workflowID, runID, reason, nil)
	if err != nil {
		c.logger.WithError(err).Error("Failed to terminate workflow")
		return fmt.Errorf("failed to terminate workflow: %w", err)
	}

	c.logger.Info("Workflow terminated successfully")
	return nil
}
