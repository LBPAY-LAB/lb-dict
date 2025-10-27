package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/lbpay-lab/conn-dict/internal/domain/aggregates"
)

// VsyncWorkflowInput represents the input for the vsync workflow
type VsyncWorkflowInput struct {
	Entry *aggregates.VsyncEntry
}

// VsyncWorkflow orchestrates the DICT vsync process
func VsyncWorkflow(ctx workflow.Context, input VsyncWorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting Vsync Workflow", "key", input.Entry.Key, "key_type", input.Entry.KeyType)

	// Configure activity options with retry policy
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Step 1: Fetch entry from DICT
	logger.Info("Fetching entry from DICT")
	var dictEntry interface{}
	err := workflow.ExecuteActivity(ctx, "FetchDictEntryActivity", input.Entry.Key, input.Entry.KeyType).Get(ctx, &dictEntry)
	if err != nil {
		logger.Error("Failed to fetch DICT entry", "error", err)
		// Mark as failed
		return workflow.ExecuteActivity(ctx, "MarkVsyncFailedActivity", input.Entry.Key, input.Entry.KeyType, err.Error()).Get(ctx, nil)
	}

	// Step 2: Validate entry data
	logger.Info("Validating entry data")
	var validationResult bool
	err = workflow.ExecuteActivity(ctx, "ValidateVsyncEntryActivity", dictEntry).Get(ctx, &validationResult)
	if err != nil || !validationResult {
		logger.Error("Entry validation failed", "error", err)
		return workflow.ExecuteActivity(ctx, "MarkVsyncFailedActivity", input.Entry.Key, input.Entry.KeyType, "Validation failed").Get(ctx, nil)
	}

	// Step 3: Sync to local database
	logger.Info("Syncing to local database")
	err = workflow.ExecuteActivity(ctx, "SyncToLocalDatabaseActivity", dictEntry).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to sync to database", "error", err)
		return workflow.ExecuteActivity(ctx, "MarkVsyncFailedActivity", input.Entry.Key, input.Entry.KeyType, err.Error()).Get(ctx, nil)
	}

	// Step 4: Publish sync event
	logger.Info("Publishing sync event")
	err = workflow.ExecuteActivity(ctx, "PublishVsyncCompletedEventActivity", input.Entry.Key, input.Entry.KeyType).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to publish event", "error", err)
		// Non-critical, continue
	}

	logger.Info("Vsync workflow completed successfully")
	return nil
}

// VsyncBatchWorkflow processes multiple vsync entries in batch
func VsyncBatchWorkflow(ctx workflow.Context, entries []*aggregates.VsyncEntry) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting Vsync Batch Workflow", "batch_size", len(entries))

	// Process entries in parallel
	var futures []workflow.Future
	for _, entry := range entries {
		childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			WorkflowID: "vsync-" + entry.KeyType + "-" + entry.Key,
		})

		future := workflow.ExecuteChildWorkflow(childCtx, VsyncWorkflow, VsyncWorkflowInput{Entry: entry})
		futures = append(futures, future)
	}

	// Wait for all to complete
	successCount := 0
	failCount := 0
	for _, future := range futures {
		err := future.Get(ctx, nil)
		if err != nil {
			logger.Error("Vsync workflow failed", "error", err)
			failCount++
		} else {
			successCount++
		}
	}

	logger.Info("Vsync batch workflow completed",
		"success", successCount,
		"failed", failCount)

	return nil
}

// ScheduledVsyncWorkflow runs periodic vsync for all entries
func ScheduledVsyncWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting Scheduled Vsync Workflow")

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Get pending entries
	var entries []*aggregates.VsyncEntry
	err := workflow.ExecuteActivity(ctx, "GetPendingVsyncEntriesActivity").Get(ctx, &entries)
	if err != nil {
		logger.Error("Failed to get pending entries", "error", err)
		return err
	}

	if len(entries) == 0 {
		logger.Info("No pending entries to sync")
		return nil
	}

	// Start batch workflow
	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID: "vsync-batch-" + workflow.Now(ctx).Format("20060102-150405"),
	})

	return workflow.ExecuteChildWorkflow(childCtx, VsyncBatchWorkflow, entries).Get(ctx, nil)
}
