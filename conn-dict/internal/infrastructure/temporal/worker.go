package temporal

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/worker"
)

// WorkerConfig holds the configuration for Temporal worker
type WorkerConfig struct {
	TaskQueue           string
	MaxConcurrentWorkflows int
	MaxConcurrentActivities int
	Logger              *logrus.Logger
}

// Worker wraps the Temporal worker with additional functionality
type Worker struct {
	worker    worker.Worker
	taskQueue string
	logger    *logrus.Logger
}

// NewWorker creates a new Temporal worker with the given configuration
func NewWorker(client *Client, cfg WorkerConfig) (*Worker, error) {
	if cfg.TaskQueue == "" {
		return nil, fmt.Errorf("task queue is required")
	}

	if cfg.MaxConcurrentWorkflows == 0 {
		cfg.MaxConcurrentWorkflows = 100
	}

	if cfg.MaxConcurrentActivities == 0 {
		cfg.MaxConcurrentActivities = 200
	}

	if cfg.Logger == nil {
		cfg.Logger = logrus.New()
		cfg.Logger.SetLevel(logrus.InfoLevel)
	}

	cfg.Logger.WithFields(logrus.Fields{
		"task_queue":                cfg.TaskQueue,
		"max_concurrent_workflows":  cfg.MaxConcurrentWorkflows,
		"max_concurrent_activities": cfg.MaxConcurrentActivities,
	}).Info("Creating Temporal worker")

	// Create worker options
	workerOptions := worker.Options{
		MaxConcurrentWorkflowTaskExecutionSize: cfg.MaxConcurrentWorkflows,
		MaxConcurrentActivityExecutionSize:     cfg.MaxConcurrentActivities,
		EnableSessionWorker:                    true,
		DisableRegistrationAliasing:            false,
	}

	// Create the worker
	w := worker.New(client.GetClient(), cfg.TaskQueue, workerOptions)

	return &Worker{
		worker:    w,
		taskQueue: cfg.TaskQueue,
		logger:    cfg.Logger,
	}, nil
}

// RegisterWorkflow registers a workflow with the worker
func (w *Worker) RegisterWorkflow(workflow interface{}) {
	w.logger.WithField("workflow", fmt.Sprintf("%T", workflow)).Info("Registering workflow")
	w.worker.RegisterWorkflow(workflow)
}

// RegisterActivity registers an activity with the worker
func (w *Worker) RegisterActivity(activity interface{}) {
	w.logger.WithField("activity", fmt.Sprintf("%T", activity)).Info("Registering activity")
	w.worker.RegisterActivity(activity)
}

// Start starts the worker
func (w *Worker) Start() error {
	w.logger.WithField("task_queue", w.taskQueue).Info("Starting Temporal worker")

	err := w.worker.Start()
	if err != nil {
		w.logger.WithError(err).Error("Failed to start worker")
		return fmt.Errorf("failed to start worker: %w", err)
	}

	w.logger.Info("Worker started successfully")
	return nil
}

// Stop stops the worker gracefully
func (w *Worker) Stop() {
	w.logger.Info("Stopping Temporal worker")
	w.worker.Stop()
	w.logger.Info("Worker stopped")
}

// Run runs the worker and blocks until context is cancelled
func (w *Worker) Run(ctx context.Context) error {
	if err := w.Start(); err != nil {
		return err
	}

	// Wait for context cancellation
	<-ctx.Done()

	w.Stop()
	return nil
}
