package workflows

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// SampleWorkflowInput defines the input for the sample workflow
type SampleWorkflowInput struct {
	Message string
}

// SampleWorkflowResult defines the result of the sample workflow
type SampleWorkflowResult struct {
	Status  string
	Message string
}

// SampleWorkflow is a simple workflow to demonstrate Temporal setup
// This is a skeleton workflow that can be used to validate the Temporal configuration
func SampleWorkflow(ctx workflow.Context, input SampleWorkflowInput) (*SampleWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("SampleWorkflow started", "input", input.Message)

	// Workflow execution logic would go here
	// This is a skeleton implementation for testing purposes

	// Sleep for a short duration to simulate work
	err := workflow.Sleep(ctx, 1*time.Second)
	if err != nil {
		return nil, err
	}

	result := &SampleWorkflowResult{
		Status:  "completed",
		Message: "Workflow executed successfully: " + input.Message,
	}

	logger.Info("SampleWorkflow completed", "result", result.Message)
	return result, nil
}
