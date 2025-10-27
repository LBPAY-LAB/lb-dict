package activities

import (
	"context"

	"go.temporal.io/sdk/activity"
)

// SampleActivityInput defines the input for the sample activity
type SampleActivityInput struct {
	Data string
}

// SampleActivityResult defines the result of the sample activity
type SampleActivityResult struct {
	Status  string
	Message string
}

// SampleActivity is a simple activity to demonstrate Temporal setup
// This is a skeleton activity that can be used to validate the Temporal configuration
func SampleActivity(ctx context.Context, input SampleActivityInput) (*SampleActivityResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("SampleActivity started", "input", input.Data)

	// Activity execution logic would go here
	// This is a skeleton implementation for testing purposes

	result := &SampleActivityResult{
		Status:  "completed",
		Message: "Activity processed: " + input.Data,
	}

	logger.Info("SampleActivity completed", "result", result.Message)
	return result, nil
}
