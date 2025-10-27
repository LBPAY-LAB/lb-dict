package workflows

import (
	"fmt"
	"time"

	"github.com/lbpay-lab/conn-dict/internal/activities"
	"go.temporal.io/sdk/workflow"
)

// VSyncInput represents the input parameters for VSYNC workflow
type VSyncInput struct {
	ParticipantISPB string     `json:"participant_ispb"` // ISPB to sync (empty = all participants)
	SyncType        string     `json:"sync_type"`        // "FULL" or "INCREMENTAL"
	LastSyncDate    *time.Time `json:"last_sync_date"`   // For incremental sync
}

// VSyncResult represents the result of the VSYNC workflow
type VSyncResult struct {
	EntriesSynced    int           `json:"entries_synced"`    // Total entries processed
	EntriesCreated   int           `json:"entries_created"`   // Missing entries created locally
	EntriesUpdated   int           `json:"entries_updated"`   // Outdated entries updated
	EntriesDeleted   int           `json:"entries_deleted"`   // Entries flagged for deletion
	Discrepancies    int           `json:"discrepancies"`     // Total discrepancies found
	Duration         time.Duration `json:"duration"`          // Workflow execution time
	SyncTimestamp    time.Time     `json:"sync_timestamp"`    // When sync started
	ReportID         string        `json:"report_id"`         // Audit report ID
	Status           string        `json:"status"`            // "COMPLETED", "PARTIAL", "FAILED"
	ErrorMessage     string        `json:"error_message,omitempty"`
}

const (
	// SyncTypeFull indicates a full synchronization (all entries)
	SyncTypeFull = "FULL"

	// SyncTypeIncremental indicates an incremental sync (only changes since last sync)
	SyncTypeIncremental = "INCREMENTAL"

	// SyncStatusCompleted indicates sync completed successfully
	SyncStatusCompleted = "COMPLETED"

	// SyncStatusPartial indicates sync completed with some errors
	SyncStatusPartial = "PARTIAL"

	// SyncStatusFailed indicates sync failed completely
	SyncStatusFailed = "FAILED"
)

// VSyncWorkflow is the main Temporal workflow for DICT data synchronization with Bacen
//
// This workflow implements periodic reconciliation between local DICT database
// and Bacen's authoritative DICT registry. It runs every 24 hours to:
// 1. Fetch entries from Bacen DICT API
// 2. Compare with local database
// 3. Detect and fix data inconsistencies
// 4. Generate audit report
// 5. Publish sync completion event
//
// Purpose: Ensure data consistency and compliance with Bacen DICT regulations
//
// Workflow Steps:
// - Step 1: Fetch Bacen entries (external API call)
// - Step 2: Compare with local database
// - Step 3: Apply fixes for discrepancies (create/update/flag)
// - Step 4: Generate audit report
// - Step 5: Publish sync event to Pulsar
func VSyncWorkflow(ctx workflow.Context, input VSyncInput) (*VSyncResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("VSyncWorkflow started",
		"ispb", input.ParticipantISPB,
		"sync_type", input.SyncType,
		"last_sync", input.LastSyncDate,
	)

	// Validate input
	if err := validateVSyncInput(input); err != nil {
		logger.Error("Invalid VSYNC input", "error", err)
		return nil, fmt.Errorf("invalid vsync input: %w", err)
	}

	startTime := workflow.Now(ctx)
	result := &VSyncResult{
		SyncTimestamp: startTime,
		Status:        SyncStatusCompleted, // Optimistic - will change if errors occur
	}

	// Get activity options
	activityOpts := activities.NewActivityOptions()

	// Step 1: Fetch entries from Bacen DICT API
	logger.Info("Step 1: Fetching entries from Bacen DICT", "ispb", input.ParticipantISPB)
	ctx1 := workflow.WithActivityOptions(ctx, activityOpts.ExternalAPI)

	var bacenEntries []activities.BacenEntry
	err := workflow.ExecuteActivity(ctx1, "FetchBacenEntriesActivity",
		input.ParticipantISPB,
		input.SyncType,
		input.LastSyncDate,
	).Get(ctx1, &bacenEntries)

	if err != nil {
		logger.Error("Failed to fetch Bacen entries", "error", err)
		result.Status = SyncStatusFailed
		result.ErrorMessage = fmt.Sprintf("Bacen API error: %v", err)
		return result, fmt.Errorf("failed to fetch Bacen entries: %w", err)
	}
	logger.Info("Fetched Bacen entries", "count", len(bacenEntries))

	// Step 2: Compare entries with local database
	logger.Info("Step 2: Comparing entries with local database")
	ctx2 := workflow.WithActivityOptions(ctx, activityOpts.Database)

	var discrepancies []activities.EntryDiscrepancy
	err = workflow.ExecuteActivity(ctx2, "CompareEntriesActivity",
		bacenEntries,
		input.ParticipantISPB,
	).Get(ctx2, &discrepancies)

	if err != nil {
		logger.Error("Failed to compare entries", "error", err)
		result.Status = SyncStatusFailed
		result.ErrorMessage = fmt.Sprintf("Comparison error: %v", err)
		return result, fmt.Errorf("failed to compare entries: %w", err)
	}

	result.Discrepancies = len(discrepancies)
	logger.Info("Found discrepancies", "count", len(discrepancies))

	// Step 3: Apply fixes for each discrepancy
	if len(discrepancies) > 0 {
		logger.Info("Step 3: Applying fixes for discrepancies")
		ctx3 := workflow.WithActivityOptions(ctx, activityOpts.Database)

		errorCount := 0
		for i, disc := range discrepancies {
			logger.Info("Processing discrepancy",
				"index", i+1,
				"total", len(discrepancies),
				"type", disc.Type,
				"key", disc.Key,
			)

			switch disc.Type {
			case activities.DiscrepancyTypeMissingLocal:
				// Entry exists in Bacen but not locally - create it
				logger.Info("Creating missing entry", "key", disc.Key)
				err = workflow.ExecuteActivity(ctx3, "CreateEntryActivity", disc.CreateInput).Get(ctx3, nil)
				if err != nil {
					logger.Warn("Failed to create missing entry", "key", disc.Key, "error", err)
					errorCount++
					continue
				}
				result.EntriesCreated++

			case activities.DiscrepancyTypeOutdatedLocal:
				// Entry exists but data is different - update it
				logger.Info("Updating outdated entry", "key", disc.Key)
				err = workflow.ExecuteActivity(ctx3, "UpdateEntryActivity", disc.EntryID, disc.UpdateInput).Get(ctx3, nil)
				if err != nil {
					logger.Warn("Failed to update entry", "key", disc.Key, "error", err)
					errorCount++
					continue
				}
				result.EntriesUpdated++

			case activities.DiscrepancyTypeMissingBacen:
				// Entry exists locally but not in Bacen - flag for manual review
				logger.Warn("Entry missing in Bacen - flagged for review",
					"key", disc.Key,
					"entry_id", disc.EntryID,
					"reason", "Entry exists locally but not found in Bacen DICT - may be pending registration or deleted",
				)
				// Don't auto-delete - could be:
				// 1. Pending registration (not yet synced to Bacen)
				// 2. Network error during Bacen fetch
				// 3. Bacen data issue
				// Flag for compliance team review instead
				result.EntriesDeleted++ // Count as flagged for deletion review

			default:
				logger.Warn("Unknown discrepancy type", "type", disc.Type, "key", disc.Key)
			}

			result.EntriesSynced++
		}

		// Set status based on error count
		if errorCount > 0 {
			result.Status = SyncStatusPartial
			result.ErrorMessage = fmt.Sprintf("%d out of %d fixes failed", errorCount, len(discrepancies))
			logger.Warn("VSYNC completed with errors", "error_count", errorCount)
		}
	} else {
		logger.Info("No discrepancies found - database is in sync")
	}

	// Step 4: Generate sync audit report
	logger.Info("Step 4: Generating sync audit report")
	ctx4 := workflow.WithActivityOptions(ctx, activityOpts.Database)

	var reportID string
	err = workflow.ExecuteActivity(ctx4, "GenerateSyncReportActivity", result).Get(ctx4, &reportID)
	if err != nil {
		logger.Warn("Failed to generate sync report (non-critical)", "error", err)
		// Don't fail workflow if report generation fails
	} else {
		result.ReportID = reportID
		logger.Info("Sync report generated", "report_id", reportID)
	}

	// Step 5: Publish sync completion event to Pulsar
	logger.Info("Step 5: Publishing sync completion event")
	ctx5 := workflow.WithActivityOptions(ctx, activityOpts.Messaging)

	syncEvent := map[string]interface{}{
		"event_type":        "vsync_completed",
		"ispb":              input.ParticipantISPB,
		"sync_type":         input.SyncType,
		"entries_synced":    result.EntriesSynced,
		"entries_created":   result.EntriesCreated,
		"entries_updated":   result.EntriesUpdated,
		"entries_deleted":   result.EntriesDeleted,
		"discrepancies":     result.Discrepancies,
		"status":            result.Status,
		"report_id":         reportID,
		"sync_timestamp":    result.SyncTimestamp,
		"duration_seconds":  workflow.Now(ctx).Sub(startTime).Seconds(),
	}

	err = workflow.ExecuteActivity(ctx5, "PublishClaimEventActivity", syncEvent).Get(ctx5, nil)
	if err != nil {
		logger.Warn("Failed to publish sync event (non-critical)", "error", err)
		// Don't fail workflow if event publishing fails
	}

	result.Duration = workflow.Now(ctx).Sub(startTime)
	logger.Info("VSyncWorkflow completed",
		"status", result.Status,
		"duration", result.Duration,
		"synced", result.EntriesSynced,
		"created", result.EntriesCreated,
		"updated", result.EntriesUpdated,
		"flagged", result.EntriesDeleted,
	)

	return result, nil
}

// validateVSyncInput validates the VSYNC workflow input
func validateVSyncInput(input VSyncInput) error {
	// SyncType must be FULL or INCREMENTAL
	if input.SyncType != SyncTypeFull && input.SyncType != SyncTypeIncremental {
		return fmt.Errorf("sync_type must be FULL or INCREMENTAL, got: %s", input.SyncType)
	}

	// For incremental sync, LastSyncDate is required
	if input.SyncType == SyncTypeIncremental && input.LastSyncDate == nil {
		return fmt.Errorf("last_sync_date is required for INCREMENTAL sync")
	}

	// ParticipantISPB is optional (empty = all participants)
	// If provided, validate format (8 digits)
	if input.ParticipantISPB != "" && len(input.ParticipantISPB) != 8 {
		return fmt.Errorf("participant_ispb must be 8 digits, got: %s", input.ParticipantISPB)
	}

	return nil
}

// VSyncSchedulerWorkflow is a cron workflow that runs VSYNC every 24 hours
//
// This workflow uses ContinueAsNew pattern to avoid workflow history bloat.
// It runs indefinitely, executing VSYNC daily and restarting itself.
//
// Scheduling:
// - Runs every 24 hours
// - Uses child workflow for actual VSYNC execution
// - Continues as new after each iteration
//
// To start the scheduler:
//   temporal workflow start \
//     --task-queue dict-task-queue \
//     --type VSyncSchedulerWorkflow \
//     --workflow-id vsync-scheduler \
//     --cron "0 2 * * *"  # Run at 2 AM daily
func VSyncSchedulerWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("VSyncSchedulerWorkflow started")

	// Execute VSYNC as child workflow
	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		WorkflowID:         fmt.Sprintf("vsync-%s", workflow.Now(ctx).Format("2006-01-02")),
		TaskQueue:          "dict-task-queue",
		WorkflowRunTimeout: 2 * time.Hour, // VSYNC should complete within 2 hours
	})

	// Configure VSYNC input
	input := VSyncInput{
		ParticipantISPB: "",                                              // Empty = all ISPBs
		SyncType:        SyncTypeIncremental,                             // Incremental by default
		LastSyncDate:    ptrTime(workflow.Now(ctx).Add(-24 * time.Hour)), // Last 24 hours
	}

	var result VSyncResult
	err := workflow.ExecuteChildWorkflow(childCtx, VSyncWorkflow, input).Get(childCtx, &result)
	if err != nil {
		logger.Error("VSYNC child workflow failed", "error", err)
		// Don't fail scheduler - continue running daily
	} else {
		logger.Info("VSYNC completed successfully",
			"status", result.Status,
			"synced", result.EntriesSynced,
			"report_id", result.ReportID,
		)
	}

	// Sleep for 24 hours
	logger.Info("Sleeping for 24 hours until next sync")
	err = workflow.Sleep(ctx, 24*time.Hour)
	if err != nil {
		return err
	}

	// Continue as new to prevent history bloat
	// This restarts the workflow with a fresh history
	logger.Info("Restarting scheduler workflow (ContinueAsNew)")
	return workflow.NewContinueAsNewError(ctx, VSyncSchedulerWorkflow)
}

// ptrTime is a helper to create *time.Time from time.Time
func ptrTime(t time.Time) *time.Time {
	return &t
}