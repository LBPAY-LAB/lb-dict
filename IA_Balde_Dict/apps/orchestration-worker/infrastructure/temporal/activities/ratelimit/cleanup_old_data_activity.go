package ratelimit

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"

	"github.com/lb-conn/connector-dict/apps/orchestration-worker/application/ports"
)

// CleanupOldDataActivity deletes states older than 13 months (retention policy)
type CleanupOldDataActivity struct {
	stateRepo ports.StateRepository
}

// CleanupResult contains the result of cleanup operation
type CleanupResult struct {
	DeletedCount int64     `json:"deleted_count"`
	CutoffDate   time.Time `json:"cutoff_date"`
}

// NewCleanupOldDataActivity creates a new CleanupOldDataActivity
func NewCleanupOldDataActivity(stateRepo ports.StateRepository) *CleanupOldDataActivity {
	return &CleanupOldDataActivity{
		stateRepo: stateRepo,
	}
}

// Execute deletes states older than 13 months
func (a *CleanupOldDataActivity) Execute(ctx context.Context) (*CleanupResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("CleanupOldDataActivity started")

	// Calculate cutoff date (13 months ago)
	now := time.Now().UTC()
	cutoffDate := now.AddDate(0, -13, 0)

	logger.Info("Deleting states older than cutoff date",
		"cutoff_date", cutoffDate.Format(time.RFC3339))

	// Delete old states
	// Note: For better performance, consider using database function drop_old_partitions()
	// instead of DELETE, as it drops entire partitions
	deletedCount, err := a.stateRepo.DeleteOlderThan(ctx, cutoffDate)
	if err != nil {
		logger.Error("Failed to delete old states", "error", err)
		return nil, fmt.Errorf("failed to delete old states: %w", err)
	}

	result := &CleanupResult{
		DeletedCount: deletedCount,
		CutoffDate:   cutoffDate,
	}

	logger.Info("CleanupOldDataActivity completed",
		"deleted_count", deletedCount,
		"cutoff_date", cutoffDate.Format(time.RFC3339))

	return result, nil
}
