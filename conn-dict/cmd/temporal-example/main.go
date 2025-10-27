package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lbpay-lab/conn-dict/internal/activities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/temporal"
	"github.com/lbpay-lab/conn-dict/internal/workflows"
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/client"
)

func main() {
	// Setup logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	logger.Info("Starting Temporal example")

	// Load configuration
	config := temporal.DefaultConfig()
	if err := config.Validate(); err != nil {
		logger.WithError(err).Fatal("Invalid configuration")
	}

	logger.WithFields(logrus.Fields{
		"host":       config.Host,
		"port":       config.Port,
		"namespace":  config.Namespace,
		"task_queue": config.TaskQueue,
	}).Info("Configuration loaded")

	// Create Temporal client
	temporalClient, err := temporal.NewClient(temporal.ClientConfig{
		HostPort:  config.HostPort(),
		Namespace: config.Namespace,
		Logger:    logger,
	})
	if err != nil {
		logger.WithError(err).Fatal("Failed to create Temporal client")
	}
	defer temporalClient.Close()

	// Perform health check
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := temporalClient.HealthCheck(ctx); err != nil {
		logger.WithError(err).Fatal("Temporal health check failed")
	}

	logger.Info("Temporal health check passed")

	// Create worker
	worker, err := temporal.NewWorker(temporalClient, temporal.WorkerConfig{
		TaskQueue:               config.TaskQueue,
		MaxConcurrentWorkflows:  100,
		MaxConcurrentActivities: 200,
		Logger:                  logger,
	})
	if err != nil {
		logger.WithError(err).Fatal("Failed to create worker")
	}

	// Register workflows and activities
	worker.RegisterWorkflow(workflows.SampleWorkflow)
	worker.RegisterActivity(activities.SampleActivity)

	logger.Info("Workflows and activities registered")

	// Start worker in a goroutine
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	go func() {
		if err := worker.Run(workerCtx); err != nil {
			logger.WithError(err).Error("Worker error")
		}
	}()

	// Wait a bit for worker to start
	time.Sleep(2 * time.Second)

	// Execute a sample workflow
	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("sample-workflow-%d", time.Now().Unix()),
		TaskQueue: config.TaskQueue,
		// Workflow execution timeout
		WorkflowExecutionTimeout: 5 * time.Minute,
		// Workflow task timeout
		WorkflowTaskTimeout: 10 * time.Second,
	}

	workflowInput := workflows.SampleWorkflowInput{
		Message: "Hello from Temporal!",
	}

	logger.Info("Executing sample workflow")

	run, err := temporalClient.ExecuteWorkflow(context.Background(), workflowOptions, workflows.SampleWorkflow, workflowInput)
	if err != nil {
		logger.WithError(err).Fatal("Failed to execute workflow")
	}

	logger.WithFields(logrus.Fields{
		"workflow_id": run.GetID(),
		"run_id":      run.GetRunID(),
	}).Info("Workflow started")

	// Wait for workflow result
	var result workflows.SampleWorkflowResult
	if err := run.Get(context.Background(), &result); err != nil {
		logger.WithError(err).Error("Failed to get workflow result")
	} else {
		logger.WithFields(logrus.Fields{
			"status":  result.Status,
			"message": result.Message,
		}).Info("Workflow completed")
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	logger.Info("Press Ctrl+C to exit")
	<-sigChan

	logger.Info("Shutting down")
	workerCancel()
	time.Sleep(1 * time.Second)
}
