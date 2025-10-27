package helpers

import (
	"testing"

	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
)

// TemporalTestEnv provides a test environment for Temporal workflows and activities
type TemporalTestEnv struct {
	Suite *testsuite.WorkflowTestSuite
	Env   *testsuite.TestWorkflowEnvironment
}

// NewTemporalTestEnv creates a new Temporal test environment
func NewTemporalTestEnv(t *testing.T) *TemporalTestEnv {
	suite := &testsuite.WorkflowTestSuite{}
	env := suite.NewTestWorkflowEnvironment()

	// Cleanup
	t.Cleanup(func() {
		// Any cleanup needed
	})

	return &TemporalTestEnv{
		Suite: suite,
		Env:   env,
	}
}

// NewTemporalActivityTestEnv creates a test environment for activities
func NewTemporalActivityTestEnv(t *testing.T) *testsuite.TestActivityEnvironment {
	suite := &testsuite.WorkflowTestSuite{}
	env := suite.NewTestActivityEnvironment()

	t.Cleanup(func() {
		// Any cleanup needed
	})

	return env
}

// RegisterWorkflow registers a workflow for testing
func (e *TemporalTestEnv) RegisterWorkflow(workflow interface{}) {
	e.Env.RegisterWorkflow(workflow)
}

// RegisterActivity registers an activity for testing
func (e *TemporalTestEnv) RegisterActivity(activity interface{}) {
	e.Env.RegisterActivity(activity)
}

// ExecuteWorkflow executes a workflow in the test environment
func (e *TemporalTestEnv) ExecuteWorkflow(workflow interface{}, args ...interface{}) {
	e.Env.ExecuteWorkflow(workflow, args...)
}

// IsWorkflowCompleted checks if the workflow has completed
func (e *TemporalTestEnv) IsWorkflowCompleted() bool {
	return e.Env.IsWorkflowCompleted()
}

// GetWorkflowError returns the workflow error if any
func (e *TemporalTestEnv) GetWorkflowError() error {
	return e.Env.GetWorkflowError()
}

// GetWorkflowResult gets the workflow result
func (e *TemporalTestEnv) GetWorkflowResult(valuePtr interface{}) error {
	return e.Env.GetWorkflowResult(valuePtr)
}

// MockActivityOptions provides options for mocking activities
type MockActivityOptions struct {
	Name   string
	Result interface{}
	Error  error
}

// SetupMockActivity sets up a mock activity
func (e *TemporalTestEnv) SetupMockActivity(opts MockActivityOptions) {
	if opts.Error != nil {
		e.Env.OnActivity(opts.Name, nil).Return(nil, opts.Error)
	} else {
		e.Env.OnActivity(opts.Name, nil).Return(opts.Result, nil)
	}
}

// WorkerTestConfig holds configuration for worker testing
type WorkerTestConfig struct {
	TaskQueue string
	Workflows []interface{}
	Activities []interface{}
}

// CreateTestWorker creates a test worker (requires Temporal server)
func CreateTestWorker(t *testing.T, config WorkerTestConfig) worker.Worker {
	// This is a placeholder for actual worker creation
	// In real tests, this would connect to a Temporal server
	t.Skip("Requires Temporal server connection")
	return nil
}
