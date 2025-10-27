package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Constructor Tests ====================

func TestNewInfraction_Success(t *testing.T) {
	// Arrange
	infractionID := "INF-12345"
	key := "+5511999999999"
	infractionType := InfractionTypeFraud
	description := "Suspected fraudulent activity on this key"
	reporterISPB := "12345678"

	// Act
	infraction, err := NewInfraction(infractionID, key, infractionType, description, reporterISPB)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, infraction)
	assert.NotEqual(t, "", infraction.ID.String())
	assert.Equal(t, infractionID, infraction.InfractionID)
	assert.Equal(t, key, infraction.Key)
	assert.Equal(t, infractionType, infraction.Type)
	assert.Equal(t, description, infraction.Description)
	assert.Equal(t, reporterISPB, infraction.ReporterParticipant)
	assert.Equal(t, InfractionStatusOpen, infraction.Status)
	assert.NotNil(t, infraction.ReportedAt)
	assert.NotNil(t, infraction.CreatedAt)
	assert.NotNil(t, infraction.UpdatedAt)
	assert.Empty(t, infraction.EvidenceURLs)
	assert.Nil(t, infraction.ResolutionNotes)
	assert.Nil(t, infraction.InvestigatedAt)
	assert.Nil(t, infraction.ResolvedAt)
}

func TestNewInfraction_InvalidReporterISPB(t *testing.T) {
	// Arrange
	testCases := []struct {
		name         string
		reporterISPB string
	}{
		{"Empty ISPB", ""},
		{"Too Short", "1234567"},
		{"Too Long", "123456789"},
		{"Non-numeric", "1234567a"},
		{"With Spaces", "1234 5678"},
	}

	// Act & Assert
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			infraction, err := NewInfraction(
				"INF-001",
				"+5511999999999",
				InfractionTypeFraud,
				"Fraud detected",
				tc.reporterISPB,
			)

			assert.Error(t, err)
			assert.Nil(t, infraction)
			assert.Contains(t, err.Error(), "invalid reporter ISPB")
		})
	}
}

func TestNewInfraction_EmptyKey(t *testing.T) {
	// Arrange
	key := ""
	reporterISPB := "12345678"

	// Act
	infraction, err := NewInfraction(
		"INF-001",
		key,
		InfractionTypeFraud,
		"Description",
		reporterISPB,
	)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, infraction)
	assert.Equal(t, "key is required", err.Error())
}

func TestNewInfraction_EmptyDescription(t *testing.T) {
	// Arrange
	description := ""
	reporterISPB := "12345678"

	// Act
	infraction, err := NewInfraction(
		"INF-001",
		"+5511999999999",
		InfractionTypeFraud,
		description,
		reporterISPB,
	)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, infraction)
	assert.Equal(t, "description is required", err.Error())
}

// ==================== Status Transition Tests ====================

func TestInfraction_Investigate_Success(t *testing.T) {
	// Arrange
	infraction, _ := NewInfraction(
		"INF-001",
		"+5511999999999",
		InfractionTypeFraud,
		"Fraud detected",
		"12345678",
	)
	assert.Equal(t, InfractionStatusOpen, infraction.Status)
	assert.Nil(t, infraction.InvestigatedAt)

	// Act
	err := infraction.Investigate()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, InfractionStatusUnderInvestigation, infraction.Status)
	assert.NotNil(t, infraction.InvestigatedAt)
	assert.True(t, infraction.UpdatedAt.After(infraction.CreatedAt))
}

func TestInfraction_Investigate_WrongStatus(t *testing.T) {
	// Arrange - create and already mark as under investigation
	infraction, _ := NewInfraction(
		"INF-001",
		"+5511999999999",
		InfractionTypeFraud,
		"Fraud detected",
		"12345678",
	)
	_ = infraction.Investigate() // Move to UNDER_INVESTIGATION

	// Act
	err := infraction.Investigate() // Try to investigate again

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "can only investigate open infractions")
	assert.Equal(t, InfractionStatusUnderInvestigation, infraction.Status)
}

func TestInfraction_Resolve_Success(t *testing.T) {
	// Arrange
	infraction, _ := NewInfraction(
		"INF-001",
		"+5511999999999",
		InfractionTypeFraud,
		"Fraud detected",
		"12345678",
	)
	_ = infraction.Investigate() // OPEN → UNDER_INVESTIGATION
	notes := "Investigation completed. No fraud found."

	// Act
	err := infraction.Resolve(notes)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, InfractionStatusResolved, infraction.Status)
	assert.NotNil(t, infraction.ResolvedAt)
	assert.NotNil(t, infraction.ResolutionNotes)
	assert.Equal(t, notes, *infraction.ResolutionNotes)
}

func TestInfraction_Resolve_WithoutNotes(t *testing.T) {
	// Arrange
	infraction, _ := NewInfraction(
		"INF-001",
		"+5511999999999",
		InfractionTypeFraud,
		"Fraud detected",
		"12345678",
	)
	_ = infraction.Investigate()

	// Act
	err := infraction.Resolve("")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "resolution notes are required", err.Error())
	assert.Equal(t, InfractionStatusUnderInvestigation, infraction.Status)
}

func TestInfraction_Resolve_FromWrongStatus(t *testing.T) {
	// Arrange - resolved infraction
	infraction, _ := NewInfraction(
		"INF-001",
		"+5511999999999",
		InfractionTypeFraud,
		"Fraud detected",
		"12345678",
	)
	_ = infraction.Investigate()
	_ = infraction.Resolve("Already resolved")

	// Act
	err := infraction.Resolve("Try to resolve again")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "can only resolve open or under investigation infractions")
	assert.Equal(t, InfractionStatusResolved, infraction.Status)
}

func TestInfraction_Dismiss_Success(t *testing.T) {
	// Arrange
	infraction, _ := NewInfraction(
		"INF-001",
		"+5511999999999",
		InfractionTypeFraud,
		"Fraud detected",
		"12345678",
	)
	notes := "False positive - dismissed"

	// Act
	err := infraction.Dismiss(notes)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, InfractionStatusDismissed, infraction.Status)
	assert.NotNil(t, infraction.ResolvedAt)
	assert.NotNil(t, infraction.ResolutionNotes)
	assert.Equal(t, notes, *infraction.ResolutionNotes)
}

func TestInfraction_Dismiss_WithoutNotes(t *testing.T) {
	// Arrange
	infraction, _ := NewInfraction(
		"INF-001",
		"+5511999999999",
		InfractionTypeFraud,
		"Fraud detected",
		"12345678",
	)

	// Act
	err := infraction.Dismiss("")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "dismissal notes are required", err.Error())
	assert.Equal(t, InfractionStatusOpen, infraction.Status)
}

func TestInfraction_Dismiss_AlreadyResolved(t *testing.T) {
	// Arrange
	infraction, _ := NewInfraction(
		"INF-001",
		"+5511999999999",
		InfractionTypeFraud,
		"Fraud detected",
		"12345678",
	)
	_ = infraction.Investigate()
	_ = infraction.Resolve("Investigation completed")

	// Act
	err := infraction.Dismiss("Try to dismiss")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot dismiss resolved or already dismissed infraction")
	assert.Equal(t, InfractionStatusResolved, infraction.Status)
}

func TestInfraction_EscalateToBacen_Success(t *testing.T) {
	// Arrange
	infraction, _ := NewInfraction(
		"INF-001",
		"+5511999999999",
		InfractionTypeFraud,
		"Serious fraud detected",
		"12345678",
	)
	notes := "Escalating to Bacen for investigation"

	// Act
	err := infraction.EscalateToBacen(notes)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, InfractionStatusEscalatedToBacen, infraction.Status)
	assert.NotNil(t, infraction.ResolutionNotes)
	assert.Equal(t, notes, *infraction.ResolutionNotes)
}

func TestInfraction_EscalateToBacen_AlreadyEscalated(t *testing.T) {
	// Arrange
	infraction, _ := NewInfraction(
		"INF-001",
		"+5511999999999",
		InfractionTypeFraud,
		"Serious fraud",
		"12345678",
	)
	_ = infraction.EscalateToBacen("First escalation")

	// Act
	err := infraction.EscalateToBacen("Try to escalate again")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "infraction already escalated to Bacen", err.Error())
	assert.Equal(t, InfractionStatusEscalatedToBacen, infraction.Status)
}

func TestInfraction_EscalateToBacen_AfterResolved(t *testing.T) {
	// Arrange
	infraction, _ := NewInfraction(
		"INF-001",
		"+5511999999999",
		InfractionTypeFraud,
		"Fraud detected",
		"12345678",
	)
	_ = infraction.Investigate()
	_ = infraction.Resolve("Already resolved")

	// Act
	err := infraction.EscalateToBacen("Try to escalate")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot escalate resolved or dismissed infraction")
	assert.Equal(t, InfractionStatusResolved, infraction.Status)
}

// ==================== Evidence Tests ====================

func TestInfraction_AddEvidence_Success(t *testing.T) {
	// Arrange
	infraction, _ := NewInfraction(
		"INF-001",
		"+5511999999999",
		InfractionTypeFraud,
		"Fraud detected",
		"12345678",
	)
	evidenceURL := "https://evidence.example.com/screenshot1.png"
	beforeUpdate := infraction.UpdatedAt

	// Sleep to ensure timestamp difference
	time.Sleep(1 * time.Millisecond)

	// Act
	err := infraction.AddEvidence(evidenceURL)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, infraction.EvidenceURLs, 1)
	assert.Equal(t, evidenceURL, infraction.EvidenceURLs[0])
	assert.True(t, infraction.UpdatedAt.After(beforeUpdate))
}

func TestInfraction_AddEvidence_EmptyURL(t *testing.T) {
	// Arrange
	infraction, _ := NewInfraction(
		"INF-001",
		"+5511999999999",
		InfractionTypeFraud,
		"Fraud detected",
		"12345678",
	)

	// Act
	err := infraction.AddEvidence("")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "evidence URL cannot be empty", err.Error())
	assert.Empty(t, infraction.EvidenceURLs)
}

func TestInfraction_AddEvidence_Duplicate(t *testing.T) {
	// Arrange
	infraction, _ := NewInfraction(
		"INF-001",
		"+5511999999999",
		InfractionTypeFraud,
		"Fraud detected",
		"12345678",
	)
	evidenceURL := "https://evidence.example.com/screenshot1.png"
	_ = infraction.AddEvidence(evidenceURL)

	// Act
	err := infraction.AddEvidence(evidenceURL) // Add same URL again

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "evidence URL already exists", err.Error())
	assert.Len(t, infraction.EvidenceURLs, 1) // Should still be 1
}

// ==================== Helper Methods Tests ====================

func TestInfraction_IsOpen_True(t *testing.T) {
	// Arrange & Act - OPEN status
	infractionOpen, _ := NewInfraction(
		"INF-001",
		"+5511999999999",
		InfractionTypeFraud,
		"Fraud detected",
		"12345678",
	)

	// Assert
	assert.True(t, infractionOpen.IsOpen())

	// Arrange & Act - UNDER_INVESTIGATION status
	infractionInvestigating, _ := NewInfraction(
		"INF-002",
		"+5511888888888",
		InfractionTypeFraud,
		"Under investigation",
		"12345678",
	)
	_ = infractionInvestigating.Investigate()

	// Assert
	assert.True(t, infractionInvestigating.IsOpen())
}

func TestInfraction_IsOpen_False_Resolved(t *testing.T) {
	// Arrange
	infraction, _ := NewInfraction(
		"INF-001",
		"+5511999999999",
		InfractionTypeFraud,
		"Fraud detected",
		"12345678",
	)
	_ = infraction.Investigate()
	_ = infraction.Resolve("Resolved")

	// Act & Assert
	assert.False(t, infraction.IsOpen())
}

func TestInfraction_IsClosed_True(t *testing.T) {
	// Arrange & Act - RESOLVED status
	infractionResolved, _ := NewInfraction(
		"INF-001",
		"+5511999999999",
		InfractionTypeFraud,
		"Fraud detected",
		"12345678",
	)
	_ = infractionResolved.Investigate()
	_ = infractionResolved.Resolve("Resolved")

	// Assert
	assert.True(t, infractionResolved.IsClosed())

	// Arrange & Act - DISMISSED status
	infractionDismissed, _ := NewInfraction(
		"INF-002",
		"+5511888888888",
		InfractionTypeFraud,
		"False positive",
		"12345678",
	)
	_ = infractionDismissed.Dismiss("Dismissed")

	// Assert
	assert.True(t, infractionDismissed.IsClosed())
}

func TestInfraction_IsClosed_False(t *testing.T) {
	// Arrange
	infraction, _ := NewInfraction(
		"INF-001",
		"+5511999999999",
		InfractionTypeFraud,
		"Fraud detected",
		"12345678",
	)

	// Act & Assert - OPEN
	assert.False(t, infraction.IsClosed())

	// Arrange - UNDER_INVESTIGATION
	_ = infraction.Investigate()

	// Act & Assert
	assert.False(t, infraction.IsClosed())
}

func TestInfraction_IsEscalated_True(t *testing.T) {
	// Arrange
	infraction, _ := NewInfraction(
		"INF-001",
		"+5511999999999",
		InfractionTypeFraud,
		"Serious fraud",
		"12345678",
	)
	_ = infraction.EscalateToBacen("Escalating to Bacen")

	// Act & Assert
	assert.True(t, infraction.IsEscalated())
}

// ==================== Validation Tests ====================

func TestInfraction_ValidateStatusTransition_Valid(t *testing.T) {
	// Arrange
	testCases := []struct {
		name        string
		fromStatus  InfractionStatus
		toStatus    InfractionStatus
		setupSteps  func(*Infraction)
	}{
		{
			name:       "OPEN to UNDER_INVESTIGATION",
			fromStatus: InfractionStatusOpen,
			toStatus:   InfractionStatusUnderInvestigation,
			setupSteps: func(i *Infraction) {
				// Already OPEN by default
			},
		},
		{
			name:       "OPEN to DISMISSED",
			fromStatus: InfractionStatusOpen,
			toStatus:   InfractionStatusDismissed,
			setupSteps: func(i *Infraction) {
				// Already OPEN by default
			},
		},
		{
			name:       "OPEN to ESCALATED_TO_BACEN",
			fromStatus: InfractionStatusOpen,
			toStatus:   InfractionStatusEscalatedToBacen,
			setupSteps: func(i *Infraction) {
				// Already OPEN by default
			},
		},
		{
			name:       "UNDER_INVESTIGATION to RESOLVED",
			fromStatus: InfractionStatusUnderInvestigation,
			toStatus:   InfractionStatusResolved,
			setupSteps: func(i *Infraction) {
				_ = i.Investigate()
			},
		},
		{
			name:       "UNDER_INVESTIGATION to DISMISSED",
			fromStatus: InfractionStatusUnderInvestigation,
			toStatus:   InfractionStatusDismissed,
			setupSteps: func(i *Infraction) {
				_ = i.Investigate()
			},
		},
		{
			name:       "ESCALATED_TO_BACEN to RESOLVED",
			fromStatus: InfractionStatusEscalatedToBacen,
			toStatus:   InfractionStatusResolved,
			setupSteps: func(i *Infraction) {
				_ = i.EscalateToBacen("Escalating")
			},
		},
	}

	// Act & Assert
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			infraction, _ := NewInfraction(
				"INF-001",
				"+5511999999999",
				InfractionTypeFraud,
				"Test",
				"12345678",
			)
			tc.setupSteps(infraction)

			err := infraction.ValidateStatusTransition(tc.toStatus)

			assert.NoError(t, err)
			assert.Equal(t, tc.fromStatus, infraction.Status)
		})
	}
}

func TestInfraction_ValidateStatusTransition_Invalid(t *testing.T) {
	// Arrange
	testCases := []struct {
		name       string
		fromStatus InfractionStatus
		toStatus   InfractionStatus
		setupSteps func(*Infraction)
	}{
		{
			name:       "OPEN to RESOLVED (must investigate first)",
			fromStatus: InfractionStatusOpen,
			toStatus:   InfractionStatusResolved,
			setupSteps: func(i *Infraction) {
				// Already OPEN by default
			},
		},
		{
			name:       "RESOLVED to DISMISSED (terminal state)",
			fromStatus: InfractionStatusResolved,
			toStatus:   InfractionStatusDismissed,
			setupSteps: func(i *Infraction) {
				_ = i.Investigate()
				_ = i.Resolve("Resolved")
			},
		},
		{
			name:       "DISMISSED to RESOLVED (terminal state)",
			fromStatus: InfractionStatusDismissed,
			toStatus:   InfractionStatusResolved,
			setupSteps: func(i *Infraction) {
				_ = i.Dismiss("Dismissed")
			},
		},
		{
			name:       "RESOLVED to UNDER_INVESTIGATION (terminal state)",
			fromStatus: InfractionStatusResolved,
			toStatus:   InfractionStatusUnderInvestigation,
			setupSteps: func(i *Infraction) {
				_ = i.Investigate()
				_ = i.Resolve("Resolved")
			},
		},
	}

	// Act & Assert
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			infraction, _ := NewInfraction(
				"INF-001",
				"+5511999999999",
				InfractionTypeFraud,
				"Test",
				"12345678",
			)
			tc.setupSteps(infraction)

			err := infraction.ValidateStatusTransition(tc.toStatus)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid status transition")
			assert.Equal(t, tc.fromStatus, infraction.Status)
		})
	}
}