package workflows

import (
	"testing"
	"time"

	"github.com/lbpay-lab/conn-dict/internal/activities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

type VSyncWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

func (s *VSyncWorkflowTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *VSyncWorkflowTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

// TestVSyncWorkflow_Success tests successful VSYNC execution
func (s *VSyncWorkflowTestSuite) TestVSyncWorkflow_Success() {
	// Arrange
	input := VSyncInput{
		ParticipantISPB: "12345678",
		SyncType:        SyncTypeFull,
		LastSyncDate:    nil,
	}

	// Mock Bacen entries
	bacenEntries := []activities.BacenEntry{
		{
			Key:             "+5511999999999",
			KeyType:         "PHONE",
			ParticipantISPB: "12345678",
			AccountBranch:   "0001",
			AccountNumber:   "123456",
			AccountType:     "CACC",
			Status:          "ACTIVE",
		},
	}

	// Mock discrepancies
	discrepancies := []activities.EntryDiscrepancy{
		{
			Type:    activities.DiscrepancyTypeMissingLocal,
			Key:     "+5511999999999",
			EntryID: "",
			CreateInput: activities.CreateEntryInput{
				EntryID:         "entry-123",
				Key:             "+5511999999999",
				KeyType:         "PHONE",
				ParticipantISPB: "12345678",
			},
		},
	}

	reportID := "SYNC-REPORT-123"

	// Mock activities
	s.env.OnActivity("FetchBacenEntriesActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(bacenEntries, nil)

	s.env.OnActivity("CompareEntriesActivity", mock.Anything, mock.Anything, mock.Anything).
		Return(discrepancies, nil)

	s.env.OnActivity("CreateEntryActivity", mock.Anything, mock.Anything).
		Return(nil)

	s.env.OnActivity("GenerateSyncReportActivity", mock.Anything, mock.Anything).
		Return(reportID, nil)

	s.env.OnActivity("PublishClaimEventActivity", mock.Anything, mock.Anything).
		Return(nil)

	// Act
	s.env.ExecuteWorkflow(VSyncWorkflow, input)

	// Assert
	assert.True(s.T(), s.env.IsWorkflowCompleted())
	assert.NoError(s.T(), s.env.GetWorkflowError())

	var result VSyncResult
	err := s.env.GetWorkflowResult(&result)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), SyncStatusCompleted, result.Status)
	assert.Equal(s.T(), 1, result.Discrepancies)
	assert.Equal(s.T(), 1, result.EntriesCreated)
	assert.Equal(s.T(), reportID, result.ReportID)
}

// TestVSyncWorkflow_NoDiscrepancies tests VSYNC when database is already in sync
func (s *VSyncWorkflowTestSuite) TestVSyncWorkflow_NoDiscrepancies() {
	// Arrange
	input := VSyncInput{
		ParticipantISPB: "",
		SyncType:        SyncTypeIncremental,
		LastSyncDate:    ptrTime(time.Now().Add(-24 * time.Hour)),
	}

	bacenEntries := []activities.BacenEntry{}
	discrepancies := []activities.EntryDiscrepancy{}
	reportID := "SYNC-REPORT-456"

	// Mock activities
	s.env.OnActivity("FetchBacenEntriesActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(bacenEntries, nil)

	s.env.OnActivity("CompareEntriesActivity", mock.Anything, mock.Anything, mock.Anything).
		Return(discrepancies, nil)

	s.env.OnActivity("GenerateSyncReportActivity", mock.Anything, mock.Anything).
		Return(reportID, nil)

	s.env.OnActivity("PublishClaimEventActivity", mock.Anything, mock.Anything).
		Return(nil)

	// Act
	s.env.ExecuteWorkflow(VSyncWorkflow, input)

	// Assert
	assert.True(s.T(), s.env.IsWorkflowCompleted())
	assert.NoError(s.T(), s.env.GetWorkflowError())

	var result VSyncResult
	err := s.env.GetWorkflowResult(&result)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), SyncStatusCompleted, result.Status)
	assert.Equal(s.T(), 0, result.Discrepancies)
	assert.Equal(s.T(), 0, result.EntriesCreated)
}

// TestVSyncWorkflow_PartialFailure tests VSYNC with some activity failures
func (s *VSyncWorkflowTestSuite) TestVSyncWorkflow_PartialFailure() {
	// Arrange
	input := VSyncInput{
		ParticipantISPB: "12345678",
		SyncType:        SyncTypeFull,
		LastSyncDate:    nil,
	}

	bacenEntries := []activities.BacenEntry{
		{Key: "+5511999999999", KeyType: "PHONE"},
		{Key: "+5511888888888", KeyType: "PHONE"},
	}

	discrepancies := []activities.EntryDiscrepancy{
		{
			Type:        activities.DiscrepancyTypeMissingLocal,
			Key:         "+5511999999999",
			CreateInput: activities.CreateEntryInput{EntryID: "entry-1", Key: "+5511999999999"},
		},
		{
			Type:        activities.DiscrepancyTypeMissingLocal,
			Key:         "+5511888888888",
			CreateInput: activities.CreateEntryInput{EntryID: "entry-2", Key: "+5511888888888"},
		},
	}

	reportID := "SYNC-REPORT-789"

	// Mock activities - second CreateEntry fails
	s.env.OnActivity("FetchBacenEntriesActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(bacenEntries, nil)

	s.env.OnActivity("CompareEntriesActivity", mock.Anything, mock.Anything, mock.Anything).
		Return(discrepancies, nil)

	s.env.OnActivity("CreateEntryActivity", mock.Anything, discrepancies[0].CreateInput).
		Return(nil)

	s.env.OnActivity("CreateEntryActivity", mock.Anything, discrepancies[1].CreateInput).
		Return(assert.AnError)

	s.env.OnActivity("GenerateSyncReportActivity", mock.Anything, mock.Anything).
		Return(reportID, nil)

	s.env.OnActivity("PublishClaimEventActivity", mock.Anything, mock.Anything).
		Return(nil)

	// Act
	s.env.ExecuteWorkflow(VSyncWorkflow, input)

	// Assert
	assert.True(s.T(), s.env.IsWorkflowCompleted())
	assert.NoError(s.T(), s.env.GetWorkflowError())

	var result VSyncResult
	err := s.env.GetWorkflowResult(&result)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), SyncStatusPartial, result.Status)
	assert.Equal(s.T(), 2, result.Discrepancies)
	assert.Equal(s.T(), 1, result.EntriesCreated) // Only first succeeded
	assert.Contains(s.T(), result.ErrorMessage, "1 out of 2 fixes failed")
}

// TestVSyncWorkflow_InvalidInput tests input validation
func (s *VSyncWorkflowTestSuite) TestVSyncWorkflow_InvalidInput() {
	// Arrange - invalid sync type
	input := VSyncInput{
		ParticipantISPB: "12345678",
		SyncType:        "INVALID",
		LastSyncDate:    nil,
	}

	// Act
	s.env.ExecuteWorkflow(VSyncWorkflow, input)

	// Assert
	assert.True(s.T(), s.env.IsWorkflowCompleted())
	assert.Error(s.T(), s.env.GetWorkflowError())
	assert.Contains(s.T(), s.env.GetWorkflowError().Error(), "sync_type must be FULL or INCREMENTAL")
}

// TestVSyncWorkflow_IncrementalWithoutDate tests incremental sync without LastSyncDate
func (s *VSyncWorkflowTestSuite) TestVSyncWorkflow_IncrementalWithoutDate() {
	// Arrange - incremental sync without LastSyncDate
	input := VSyncInput{
		ParticipantISPB: "12345678",
		SyncType:        SyncTypeIncremental,
		LastSyncDate:    nil, // Missing
	}

	// Act
	s.env.ExecuteWorkflow(VSyncWorkflow, input)

	// Assert
	assert.True(s.T(), s.env.IsWorkflowCompleted())
	assert.Error(s.T(), s.env.GetWorkflowError())
	assert.Contains(s.T(), s.env.GetWorkflowError().Error(), "last_sync_date is required")
}

// TestVSyncWorkflow_MultipleDiscrepancyTypes tests handling of different discrepancy types
func (s *VSyncWorkflowTestSuite) TestVSyncWorkflow_MultipleDiscrepancyTypes() {
	// Arrange
	input := VSyncInput{
		ParticipantISPB: "12345678",
		SyncType:        SyncTypeFull,
		LastSyncDate:    nil,
	}

	bacenEntries := []activities.BacenEntry{
		{Key: "+5511999999999"},
	}

	discrepancies := []activities.EntryDiscrepancy{
		{
			Type:        activities.DiscrepancyTypeMissingLocal,
			Key:         "+5511999999999",
			CreateInput: activities.CreateEntryInput{EntryID: "entry-1"},
		},
		{
			Type:        activities.DiscrepancyTypeOutdatedLocal,
			Key:         "+5511888888888",
			EntryID:     "entry-2",
			UpdateInput: activities.UpdateEntryInput{},
		},
		{
			Type:    activities.DiscrepancyTypeMissingBacen,
			Key:     "+5511777777777",
			EntryID: "entry-3",
		},
	}

	reportID := "SYNC-REPORT-999"

	// Mock activities
	s.env.OnActivity("FetchBacenEntriesActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(bacenEntries, nil)

	s.env.OnActivity("CompareEntriesActivity", mock.Anything, mock.Anything, mock.Anything).
		Return(discrepancies, nil)

	s.env.OnActivity("CreateEntryActivity", mock.Anything, mock.Anything).
		Return(nil)

	s.env.OnActivity("UpdateEntryActivity", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	s.env.OnActivity("GenerateSyncReportActivity", mock.Anything, mock.Anything).
		Return(reportID, nil)

	s.env.OnActivity("PublishClaimEventActivity", mock.Anything, mock.Anything).
		Return(nil)

	// Act
	s.env.ExecuteWorkflow(VSyncWorkflow, input)

	// Assert
	assert.True(s.T(), s.env.IsWorkflowCompleted())
	assert.NoError(s.T(), s.env.GetWorkflowError())

	var result VSyncResult
	err := s.env.GetWorkflowResult(&result)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), SyncStatusCompleted, result.Status)
	assert.Equal(s.T(), 3, result.Discrepancies)
	assert.Equal(s.T(), 1, result.EntriesCreated)
	assert.Equal(s.T(), 1, result.EntriesUpdated)
	assert.Equal(s.T(), 1, result.EntriesDeleted) // Flagged for deletion review
}

func TestVSyncWorkflowTestSuite(t *testing.T) {
	suite.Run(t, new(VSyncWorkflowTestSuite))
}