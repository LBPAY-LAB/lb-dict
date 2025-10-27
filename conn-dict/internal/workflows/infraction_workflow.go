package workflows

import (
	"fmt"
	"time"

	"github.com/lbpay-lab/conn-dict/internal/activities"
	"go.temporal.io/sdk/workflow"
)

// InvestigateInfractionInput represents the input parameters for the Infraction Investigation workflow
type InvestigateInfractionInput struct {
	InfractionID        string   `json:"infraction_id"`
	Key                 string   `json:"key"`
	Type                string   `json:"type"` // FRAUD, ACCOUNT_CLOSED, INCORRECT_DATA, UNAUTHORIZED_USE, DUPLICATE_KEY, OTHER
	Description         string   `json:"description"`
	ReporterISPB        string   `json:"reporter_ispb"`
	ReportedISPB        string   `json:"reported_ispb"`
	EvidenceURLs        []string `json:"evidence_urls,omitempty"`
	RelatedEntryID      string   `json:"related_entry_id,omitempty"`
	RelatedClaimID      string   `json:"related_claim_id,omitempty"`
}

// InvestigationDecision represents the decision made after investigation
type InvestigationDecision struct {
	Decision string `json:"decision"` // "RESOLVE", "DISMISS", "ESCALATE"
	Notes    string `json:"notes"`
}

// EvidenceData represents evidence added during investigation
type EvidenceData struct {
	EvidenceURL string `json:"evidence_url"`
	UploadedBy  string `json:"uploaded_by,omitempty"`
	Description string `json:"description,omitempty"`
}

// InfractionWorkflowResult represents the result of the Infraction workflow
type InfractionWorkflowResult struct {
	InfractionID      string    `json:"infraction_id"`
	Status            string    `json:"status"` // "RESOLVED", "DISMISSED", "ESCALATED"
	Decision          string    `json:"decision"`
	CompletedAt       time.Time `json:"completed_at"`
	ResolutionNotes   string    `json:"resolution_notes,omitempty"`
	EscalatedToBacen  bool      `json:"escalated_to_bacen"`
	Message           string    `json:"message"`
}

const (
	// InfractionTimeout is the maximum duration for an infraction investigation (30 days)
	InfractionTimeout = 30 * 24 * time.Hour

	// InvestigationTimeout is the auto-escalation timeout (7 days)
	InvestigationTimeout = 7 * 24 * time.Hour

	// InfractionStatusOpen indicates the infraction is newly created
	InfractionStatusOpen = "OPEN"

	// InfractionStatusUnderInvestigation indicates the infraction is being investigated
	InfractionStatusUnderInvestigation = "UNDER_INVESTIGATION"

	// InfractionStatusResolved indicates the infraction was resolved
	InfractionStatusResolved = "RESOLVED"

	// InfractionStatusDismissed indicates the infraction was dismissed
	InfractionStatusDismissed = "DISMISSED"

	// InfractionStatusEscalated indicates the infraction was escalated to Bacen
	InfractionStatusEscalated = "ESCALATED_TO_BACEN"
)

// InvestigateInfractionWorkflow is the main Temporal workflow for handling DICT infraction investigations
//
// This workflow implements the infraction investigation process defined by Bacen:
// 1. Infraction is created (OPEN)
// 2. Reported participant is notified
// 3. Investigation begins (UNDER_INVESTIGATION)
// 4. Evidence can be added via signals during investigation
// 5. Wait for investigation decision or 7-day timeout (auto-escalation)
// 6. Three possible outcomes:
//    a) RESOLVE → Investigation resolved with resolution notes
//    b) DISMISS → Infraction dismissed (unfounded/invalid)
//    c) ESCALATE → Escalated to Bacen for further action
// 7. Final event published and workflow completes
//
// Signals:
// - "evidence_added" → Adds evidence URL to the infraction
// - "investigation_complete" → Provides investigation decision (RESOLVE/DISMISS/ESCALATE)
//
// Timeouts:
// - Total workflow: 30 days (infractions expire after 30 days)
// - Investigation timeout: 7 days before auto-escalation to Bacen
func InvestigateInfractionWorkflow(ctx workflow.Context, input InvestigateInfractionInput) (*InfractionWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("InvestigateInfractionWorkflow started",
		"infraction_id", input.InfractionID,
		"key", input.Key,
		"type", input.Type,
		"reporter_ispb", input.ReporterISPB,
		"reported_ispb", input.ReportedISPB,
	)

	// Validate input
	if err := validateInfractionInput(input); err != nil {
		return nil, fmt.Errorf("invalid infraction input: %w", err)
	}

	// Get standardized activity options
	activityOpts := activities.NewActivityOptions()

	// Result object to track workflow outcome
	result := &InfractionWorkflowResult{
		InfractionID: input.InfractionID,
	}

	// Step 1: Create infraction in database (OPEN status)
	logger.Info("Step 1: Creating infraction in database", "infraction_id", input.InfractionID)
	ctx1 := workflow.WithActivityOptions(ctx, activityOpts.Database)
	createInput := activities.CreateInfractionInput{
		InfractionID:        input.InfractionID,
		Key:                 input.Key,
		Type:                input.Type,
		Description:         input.Description,
		ReporterParticipant: input.ReporterISPB,
		ReportedParticipant: input.ReportedISPB,
		EvidenceURLs:        input.EvidenceURLs,
		EntryID:             input.RelatedEntryID,
		ClaimID:             input.RelatedClaimID,
	}
	err := workflow.ExecuteActivity(ctx1, "CreateInfractionActivity", createInput).Get(ctx1, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create infraction: %w", err)
	}

	// Step 2: Notify reported participant (non-critical)
	logger.Info("Step 2: Notifying reported participant", "reported_ispb", input.ReportedISPB)
	ctx2 := workflow.WithActivityOptions(ctx, activityOpts.ExternalAPI)
	err = workflow.ExecuteActivity(ctx2, "NotifyReportedParticipantActivity", input.InfractionID).Get(ctx2, nil)
	if err != nil {
		logger.Warn("Failed to notify reported participant (non-critical)", "error", err)
		// Non-critical error, continue workflow
	}

	// Step 3: Wait for manual investigation start signal or auto-start after 1 hour
	logger.Info("Step 3: Waiting for investigation to begin...")

	// For this implementation, we'll auto-start investigation immediately
	// In production, you could wait for a "start_investigation" signal
	investigationStartTimer := workflow.NewTimer(ctx, 1*time.Second)
	investigationStartTimer.Get(ctx, nil) // Wait for timer

	// Step 4: Mark infraction as UNDER_INVESTIGATION
	logger.Info("Step 4: Marking infraction as under investigation", "infraction_id", input.InfractionID)
	ctx4 := workflow.WithActivityOptions(ctx, activityOpts.Database)
	err = workflow.ExecuteActivity(ctx4, "InvestigateInfractionActivity", input.InfractionID).Get(ctx4, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to mark infraction as under investigation: %w", err)
	}

	// Step 5: Handle evidence additions and wait for investigation decision
	logger.Info("Step 5: Waiting for evidence and investigation decision (7 days timeout)...")

	// Create signal channels
	evidenceChannel := workflow.GetSignalChannel(ctx, "evidence_added")
	investigationCompleteChannel := workflow.GetSignalChannel(ctx, "investigation_complete")

	var decision InvestigationDecision
	decisionMade := false

	// Create selector to handle multiple signals and timeout
	selector := workflow.NewSelector(ctx)

	// Handle evidence_added signal (can be received multiple times)
	selector.AddReceive(evidenceChannel, func(c workflow.ReceiveChannel, more bool) {
		var evidenceData EvidenceData
		c.Receive(ctx, &evidenceData)

		logger.Info("Evidence received",
			"infraction_id", input.InfractionID,
			"evidence_url", evidenceData.EvidenceURL,
		)

		// Add evidence to infraction
		ctx5 := workflow.WithActivityOptions(ctx, activityOpts.Database)
		err := workflow.ExecuteActivity(ctx5, "AddEvidenceActivity", input.InfractionID, evidenceData.EvidenceURL).Get(ctx5, nil)
		if err != nil {
			logger.Error("Failed to add evidence", "error", err)
		} else {
			logger.Info("Evidence added successfully", "evidence_url", evidenceData.EvidenceURL)
		}
	})

	// Handle investigation_complete signal (received once)
	selector.AddReceive(investigationCompleteChannel, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &decision)
		logger.Info("Investigation decision received",
			"infraction_id", input.InfractionID,
			"decision", decision.Decision,
		)
		decisionMade = true
	})

	// Handle 7-day timeout (auto-escalation)
	investigationTimer := workflow.NewTimer(ctx, InvestigationTimeout)
	selector.AddFuture(investigationTimer, func(f workflow.Future) {
		logger.Warn("Investigation timeout reached - auto-escalating to Bacen",
			"infraction_id", input.InfractionID,
			"timeout", InvestigationTimeout,
		)
		decision.Decision = "ESCALATE"
		decision.Notes = fmt.Sprintf("Auto-escalated after %s without decision", InvestigationTimeout)
		decisionMade = true
	})

	// Loop to handle multiple evidence signals until decision is made
	for !decisionMade {
		selector.Select(ctx)

		// If decision was made, break the loop
		if decisionMade {
			break
		}
	}

	logger.Info("Investigation decision finalized",
		"infraction_id", input.InfractionID,
		"decision", decision.Decision,
		"notes", decision.Notes,
	)

	// Step 6: Execute decision based on investigation outcome
	logger.Info("Step 6: Executing investigation decision", "decision", decision.Decision)
	ctx6 := workflow.WithActivityOptions(ctx, activityOpts.Database)

	switch decision.Decision {
	case "RESOLVE":
		logger.Info("Resolving infraction", "infraction_id", input.InfractionID)
		err = workflow.ExecuteActivity(ctx6, "ResolveInfractionActivity", input.InfractionID, decision.Notes).Get(ctx6, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve infraction: %w", err)
		}

		result.Status = InfractionStatusResolved
		result.Decision = "RESOLVE"
		result.ResolutionNotes = decision.Notes
		result.EscalatedToBacen = false
		result.Message = "Infraction resolved successfully"

		// Notify Bacen about resolution (non-critical)
		ctx7 := workflow.WithActivityOptions(ctx, activityOpts.ExternalAPI)
		err = workflow.ExecuteActivity(ctx7, "NotifyBacenActivity", input.InfractionID).Get(ctx7, nil)
		if err != nil {
			logger.Warn("Failed to notify Bacen about resolution (non-critical)", "error", err)
		}

	case "DISMISS":
		logger.Info("Dismissing infraction", "infraction_id", input.InfractionID)
		err = workflow.ExecuteActivity(ctx6, "DismissInfractionActivity", input.InfractionID, decision.Notes).Get(ctx6, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to dismiss infraction: %w", err)
		}

		result.Status = InfractionStatusDismissed
		result.Decision = "DISMISS"
		result.ResolutionNotes = decision.Notes
		result.EscalatedToBacen = false
		result.Message = "Infraction dismissed - unfounded or invalid"

	case "ESCALATE":
		logger.Info("Escalating infraction to Bacen", "infraction_id", input.InfractionID)
		err = workflow.ExecuteActivity(ctx6, "EscalateInfractionActivity", input.InfractionID, decision.Notes).Get(ctx6, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to escalate infraction: %w", err)
		}

		result.Status = InfractionStatusEscalated
		result.Decision = "ESCALATE"
		result.ResolutionNotes = decision.Notes
		result.EscalatedToBacen = true
		result.Message = "Infraction escalated to Bacen for further action"

		// Notify Bacen about escalation (critical)
		ctx7 := workflow.WithActivityOptions(ctx, activityOpts.ExternalAPI)
		err = workflow.ExecuteActivity(ctx7, "NotifyBacenActivity", input.InfractionID).Get(ctx7, nil)
		if err != nil {
			logger.Error("Failed to notify Bacen about escalation", "error", err)
			// For ESCALATE decision, Bacen notification is critical
			return nil, fmt.Errorf("failed to notify Bacen: %w", err)
		}

	default:
		return nil, fmt.Errorf("invalid investigation decision: %s (must be RESOLVE, DISMISS, or ESCALATE)", decision.Decision)
	}

	result.CompletedAt = workflow.Now(ctx)

	// Step 7: Publish final infraction event
	logger.Info("Step 7: Publishing final infraction event")
	ctx8 := workflow.WithActivityOptions(ctx, activityOpts.Messaging)

	finalEvent := map[string]interface{}{
		"event_type":        "infraction_workflow_completed",
		"infraction_id":     result.InfractionID,
		"key":               input.Key,
		"final_status":      result.Status,
		"decision":          result.Decision,
		"resolution_notes":  result.ResolutionNotes,
		"escalated_to_bacen": result.EscalatedToBacen,
		"completed_at":      result.CompletedAt,
	}

	err = workflow.ExecuteActivity(ctx8, "PublishInfractionEventActivity", finalEvent).Get(ctx8, nil)
	if err != nil {
		logger.Warn("Failed to publish final infraction event (non-critical)", "error", err)
	}

	logger.Info("InvestigateInfractionWorkflow completed successfully",
		"infraction_id", input.InfractionID,
		"status", result.Status,
		"decision", result.Decision,
		"escalated_to_bacen", result.EscalatedToBacen,
	)

	return result, nil
}

// validateInfractionInput validates the infraction workflow input
func validateInfractionInput(input InvestigateInfractionInput) error {
	if input.InfractionID == "" {
		return fmt.Errorf("infraction_id is required")
	}
	if input.Key == "" {
		return fmt.Errorf("key is required")
	}
	if input.Type == "" {
		return fmt.Errorf("type is required")
	}

	// Validate infraction type
	validTypes := map[string]bool{
		"FRAUD":             true,
		"ACCOUNT_CLOSED":    true,
		"INCORRECT_DATA":    true,
		"UNAUTHORIZED_USE":  true,
		"DUPLICATE_KEY":     true,
		"OTHER":             true,
	}
	if !validTypes[input.Type] {
		return fmt.Errorf("invalid infraction type: %s (must be FRAUD, ACCOUNT_CLOSED, INCORRECT_DATA, UNAUTHORIZED_USE, DUPLICATE_KEY, or OTHER)", input.Type)
	}

	if input.Description == "" {
		return fmt.Errorf("description is required")
	}
	if input.ReporterISPB == "" {
		return fmt.Errorf("reporter_ispb is required")
	}

	// Reporter and reported participant must be different
	if input.ReportedISPB != "" && input.ReporterISPB == input.ReportedISPB {
		return fmt.Errorf("reporter_ispb and reported_ispb must be different")
	}

	return nil
}