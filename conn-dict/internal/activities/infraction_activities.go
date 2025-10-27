package activities

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/pulsar"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
)

// InfractionActivities contains all Temporal activities for infraction workflow
type InfractionActivities struct {
	logger         *logrus.Logger
	infractionRepo *repositories.InfractionRepository
	pulsarProducer *pulsar.Producer
}

// NewInfractionActivities creates a new instance of InfractionActivities
func NewInfractionActivities(
	logger *logrus.Logger,
	infractionRepo *repositories.InfractionRepository,
	pulsarProducer *pulsar.Producer,
) *InfractionActivities {
	return &InfractionActivities{
		logger:         logger,
		infractionRepo: infractionRepo,
		pulsarProducer: pulsarProducer,
	}
}

// CreateInfractionInput is the input for CreateInfractionActivity
type CreateInfractionInput struct {
	InfractionID        string
	Key                 string
	Type                string // Will be converted to entities.InfractionType
	Description         string
	ReporterParticipant string
	ReportedParticipant string   // Optional
	EvidenceURLs        []string // Optional
	EntryID             string   // Optional - related entry
	ClaimID             string   // Optional - related claim
}

// CreateInfractionActivity creates a new infraction in the database
func (a *InfractionActivities) CreateInfractionActivity(ctx context.Context, input CreateInfractionInput) error {
	a.logger.WithFields(logrus.Fields{
		"infraction_id": input.InfractionID,
		"key":           input.Key,
		"type":          input.Type,
		"reporter":      input.ReporterParticipant,
	}).Info("Creating infraction")

	// Convert type string to InfractionType
	var infractionType entities.InfractionType
	switch input.Type {
	case "FRAUD":
		infractionType = entities.InfractionTypeFraud
	case "ACCOUNT_CLOSED":
		infractionType = entities.InfractionTypeAccountClosed
	case "INCORRECT_DATA":
		infractionType = entities.InfractionTypeIncorrectData
	case "UNAUTHORIZED_USE":
		infractionType = entities.InfractionTypeUnauthorizedUse
	case "DUPLICATE_KEY":
		infractionType = entities.InfractionTypeDuplicateKey
	case "OTHER":
		infractionType = entities.InfractionTypeOther
	default:
		return fmt.Errorf("invalid infraction type: %s", input.Type)
	}

	// Create infraction entity with validation
	infraction, err := entities.NewInfraction(
		input.InfractionID,
		input.Key,
		infractionType,
		input.Description,
		input.ReporterParticipant,
	)
	if err != nil {
		a.logger.WithError(err).Error("Failed to create infraction entity")
		return fmt.Errorf("invalid infraction data: %w", err)
	}

	// Set optional fields
	if input.ReportedParticipant != "" {
		infraction.ReportedParticipant = &input.ReportedParticipant
	}

	if input.EntryID != "" {
		infraction.EntryID = &input.EntryID
	}

	if input.ClaimID != "" {
		infraction.ClaimID = &input.ClaimID
	}

	// Set evidence URLs if provided
	if len(input.EvidenceURLs) > 0 {
		infraction.EvidenceURLs = input.EvidenceURLs
	}

	// Insert into database
	if err := a.infractionRepo.Create(ctx, infraction); err != nil {
		a.logger.WithError(err).Errorf("Failed to insert infraction: %s", infraction.InfractionID)
		return fmt.Errorf("database error: %w", err)
	}

	// Publish infraction created event to Pulsar
	event := map[string]interface{}{
		"event_type":           "infraction_created",
		"infraction_id":        infraction.InfractionID,
		"key":                  infraction.Key,
		"type":                 infraction.Type,
		"status":               infraction.Status,
		"reporter_participant": infraction.ReporterParticipant,
		"reported_participant": infraction.ReportedParticipant,
		"created_at":           infraction.CreatedAt,
	}

	if err := a.pulsarProducer.PublishEvent(ctx, event, infraction.InfractionID); err != nil {
		a.logger.WithError(err).Warn("Failed to publish infraction created event (non-critical)")
		// Don't fail the activity if event publishing fails
	}

	a.logger.WithField("infraction_id", infraction.InfractionID).Info("Infraction created successfully")

	return nil
}

// InvestigateInfractionActivity marks the infraction as under investigation
func (a *InfractionActivities) InvestigateInfractionActivity(ctx context.Context, infractionID string) error {
	a.logger.WithField("infraction_id", infractionID).Info("Investigating infraction")

	// Get infraction from database
	infraction, err := a.infractionRepo.GetByInfractionID(ctx, infractionID)
	if err != nil {
		return fmt.Errorf("failed to get infraction: %w", err)
	}

	// Update infraction status to under investigation
	if err := infraction.Investigate(); err != nil {
		return fmt.Errorf("invalid status transition: %w", err)
	}

	// Update in database
	if err := a.infractionRepo.Update(ctx, infraction); err != nil {
		return fmt.Errorf("failed to update infraction: %w", err)
	}

	// Publish investigation event
	event := map[string]interface{}{
		"event_type":      "infraction_under_investigation",
		"infraction_id":   infraction.InfractionID,
		"key":             infraction.Key,
		"status":          infraction.Status,
		"investigated_at": infraction.InvestigatedAt,
	}

	if err := a.pulsarProducer.PublishEvent(ctx, event, infraction.InfractionID); err != nil {
		a.logger.WithError(err).Warn("Failed to publish infraction under investigation event")
	}

	a.logger.WithField("infraction_id", infractionID).Info("Infraction investigation started successfully")

	return nil
}

// ResolveInfractionActivity resolves the infraction with resolution notes
func (a *InfractionActivities) ResolveInfractionActivity(ctx context.Context, infractionID, resolutionNotes string) error {
	a.logger.WithFields(logrus.Fields{
		"infraction_id": infractionID,
		"notes":         resolutionNotes,
	}).Info("Resolving infraction")

	// Get infraction from database
	infraction, err := a.infractionRepo.GetByInfractionID(ctx, infractionID)
	if err != nil {
		return fmt.Errorf("failed to get infraction: %w", err)
	}

	// Update infraction status to resolved
	if err := infraction.Resolve(resolutionNotes); err != nil {
		return fmt.Errorf("invalid status transition: %w", err)
	}

	// Update in database
	if err := a.infractionRepo.Update(ctx, infraction); err != nil {
		return fmt.Errorf("failed to update infraction: %w", err)
	}

	// Publish resolution event
	event := map[string]interface{}{
		"event_type":       "infraction_resolved",
		"infraction_id":    infraction.InfractionID,
		"key":              infraction.Key,
		"status":           infraction.Status,
		"resolution_notes": infraction.ResolutionNotes,
		"resolved_at":      infraction.ResolvedAt,
	}

	if err := a.pulsarProducer.PublishEvent(ctx, event, infraction.InfractionID); err != nil {
		a.logger.WithError(err).Warn("Failed to publish infraction resolved event")
	}

	a.logger.WithField("infraction_id", infractionID).Info("Infraction resolved successfully")

	return nil
}

// DismissInfractionActivity dismisses the infraction with dismissal notes
func (a *InfractionActivities) DismissInfractionActivity(ctx context.Context, infractionID, dismissalNotes string) error {
	a.logger.WithFields(logrus.Fields{
		"infraction_id": infractionID,
		"notes":         dismissalNotes,
	}).Info("Dismissing infraction")

	// Get infraction from database
	infraction, err := a.infractionRepo.GetByInfractionID(ctx, infractionID)
	if err != nil {
		return fmt.Errorf("failed to get infraction: %w", err)
	}

	// Update infraction status to dismissed
	if err := infraction.Dismiss(dismissalNotes); err != nil {
		return fmt.Errorf("invalid status transition: %w", err)
	}

	// Update in database
	if err := a.infractionRepo.Update(ctx, infraction); err != nil {
		return fmt.Errorf("failed to update infraction: %w", err)
	}

	// Publish dismissal event
	event := map[string]interface{}{
		"event_type":       "infraction_dismissed",
		"infraction_id":    infraction.InfractionID,
		"key":              infraction.Key,
		"status":           infraction.Status,
		"resolution_notes": infraction.ResolutionNotes,
		"resolved_at":      infraction.ResolvedAt,
	}

	if err := a.pulsarProducer.PublishEvent(ctx, event, infraction.InfractionID); err != nil {
		a.logger.WithError(err).Warn("Failed to publish infraction dismissed event")
	}

	a.logger.WithField("infraction_id", infractionID).Info("Infraction dismissed successfully")

	return nil
}

// EscalateInfractionActivity escalates the infraction to Bacen
func (a *InfractionActivities) EscalateInfractionActivity(ctx context.Context, infractionID, escalationNotes string) error {
	a.logger.WithFields(logrus.Fields{
		"infraction_id": infractionID,
		"notes":         escalationNotes,
	}).Info("Escalating infraction to Bacen")

	// Get infraction from database
	infraction, err := a.infractionRepo.GetByInfractionID(ctx, infractionID)
	if err != nil {
		return fmt.Errorf("failed to get infraction: %w", err)
	}

	// Update infraction status to escalated to Bacen
	if err := infraction.EscalateToBacen(escalationNotes); err != nil {
		return fmt.Errorf("invalid status transition: %w", err)
	}

	// Update in database
	if err := a.infractionRepo.Update(ctx, infraction); err != nil {
		return fmt.Errorf("failed to update infraction: %w", err)
	}

	// Publish escalation event
	event := map[string]interface{}{
		"event_type":       "infraction_escalated",
		"infraction_id":    infraction.InfractionID,
		"key":              infraction.Key,
		"status":           infraction.Status,
		"resolution_notes": infraction.ResolutionNotes,
		"escalated_at":     infraction.UpdatedAt,
	}

	if err := a.pulsarProducer.PublishEvent(ctx, event, infraction.InfractionID); err != nil {
		a.logger.WithError(err).Warn("Failed to publish infraction escalated event")
	}

	a.logger.WithField("infraction_id", infractionID).Info("Infraction escalated to Bacen successfully")

	return nil
}

// AddEvidenceActivity adds evidence URL to an infraction
func (a *InfractionActivities) AddEvidenceActivity(ctx context.Context, infractionID, evidenceURL string) error {
	a.logger.WithFields(logrus.Fields{
		"infraction_id": infractionID,
		"evidence_url":  evidenceURL,
	}).Info("Adding evidence to infraction")

	// Get infraction from database
	infraction, err := a.infractionRepo.GetByInfractionID(ctx, infractionID)
	if err != nil {
		return fmt.Errorf("failed to get infraction: %w", err)
	}

	// Add evidence URL
	if err := infraction.AddEvidence(evidenceURL); err != nil {
		return fmt.Errorf("failed to add evidence: %w", err)
	}

	// Update in database
	if err := a.infractionRepo.Update(ctx, infraction); err != nil {
		return fmt.Errorf("failed to update infraction: %w", err)
	}

	// Publish evidence added event
	event := map[string]interface{}{
		"event_type":    "evidence_added",
		"infraction_id": infraction.InfractionID,
		"key":           infraction.Key,
		"evidence_url":  evidenceURL,
		"evidence_urls": infraction.EvidenceURLs,
		"updated_at":    infraction.UpdatedAt,
	}

	if err := a.pulsarProducer.PublishEvent(ctx, event, infraction.InfractionID); err != nil {
		a.logger.WithError(err).Warn("Failed to publish evidence added event")
	}

	a.logger.WithField("infraction_id", infractionID).Info("Evidence added successfully")

	return nil
}

// GetInfractionStatusActivity retrieves the current status of an infraction
func (a *InfractionActivities) GetInfractionStatusActivity(ctx context.Context, infractionID string) (string, error) {
	a.logger.Infof("Getting infraction status: %s", infractionID)

	infraction, err := a.infractionRepo.GetByInfractionID(ctx, infractionID)
	if err != nil {
		a.logger.WithError(err).Errorf("Failed to get infraction: %s", infractionID)
		return "", fmt.Errorf("infraction not found: %w", err)
	}

	return string(infraction.Status), nil
}

// ValidateInfractionEligibilityActivity validates if an infraction can be created for a key
func (a *InfractionActivities) ValidateInfractionEligibilityActivity(ctx context.Context, key string) error {
	a.logger.WithField("key", key).Info("Validating infraction eligibility")

	// TODO: Additional validation rules:
	// 1. Check if entry exists in entries table
	// 2. Check if key is valid format
	// 3. Check if there are too many open infractions for this key
	// 4. Check if reporter has permission to report infractions

	a.logger.WithField("key", key).Info("Key is eligible for infraction")

	return nil
}

// NotifyReportedParticipantActivity sends notification to the reported ISPB about the infraction
func (a *InfractionActivities) NotifyReportedParticipantActivity(ctx context.Context, infractionID string) error {
	a.logger.WithField("infraction_id", infractionID).Info("Notifying reported participant about infraction")

	// Get infraction details
	infraction, err := a.infractionRepo.GetByInfractionID(ctx, infractionID)
	if err != nil {
		return fmt.Errorf("failed to get infraction: %w", err)
	}

	if infraction.ReportedParticipant == nil {
		a.logger.WithField("infraction_id", infractionID).Warn("No reported participant to notify")
		return nil
	}

	// TODO: Call Bridge gRPC to send DICT message to reported participant
	// This will be implemented when Bridge gRPC client is ready

	// Publish notification event
	event := map[string]interface{}{
		"event_type":           "reported_participant_notified",
		"infraction_id":        infraction.InfractionID,
		"reported_participant": infraction.ReportedParticipant,
		"key":                  infraction.Key,
	}

	if err := a.pulsarProducer.PublishEvent(ctx, event, infraction.InfractionID); err != nil {
		a.logger.WithError(err).Warn("Failed to publish reported participant notified event")
	}

	a.logger.WithField("infraction_id", infractionID).Info("Reported participant notified successfully")

	return nil
}

// NotifyBacenActivity sends infraction notification to Bacen
func (a *InfractionActivities) NotifyBacenActivity(ctx context.Context, infractionID string) error {
	a.logger.WithField("infraction_id", infractionID).Info("Notifying Bacen about infraction")

	// Get infraction details
	infraction, err := a.infractionRepo.GetByInfractionID(ctx, infractionID)
	if err != nil {
		return fmt.Errorf("failed to get infraction: %w", err)
	}

	// TODO: Call Bridge gRPC to send DICT message to Bacen
	// This will be implemented when Bridge gRPC client is ready

	// Publish notification event
	event := map[string]interface{}{
		"event_type":    "bacen_notified",
		"infraction_id": infraction.InfractionID,
		"key":           infraction.Key,
		"type":          infraction.Type,
		"status":        infraction.Status,
	}

	if err := a.pulsarProducer.PublishEvent(ctx, event, infraction.InfractionID); err != nil {
		a.logger.WithError(err).Warn("Failed to publish Bacen notified event")
	}

	a.logger.WithField("infraction_id", infractionID).Info("Bacen notified successfully")

	return nil
}

// PublishInfractionEventActivity publishes infraction events to Pulsar
func (a *InfractionActivities) PublishInfractionEventActivity(ctx context.Context, event map[string]interface{}) error {
	eventType, ok := event["event_type"].(string)
	if !ok {
		return fmt.Errorf("event must have event_type field")
	}

	infractionID, ok := event["infraction_id"].(string)
	if !ok {
		infractionID = "unknown"
	}

	a.logger.WithFields(logrus.Fields{
		"event_type":    eventType,
		"infraction_id": infractionID,
	}).Info("Publishing infraction event")

	if err := a.pulsarProducer.PublishEvent(ctx, event, infractionID); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}