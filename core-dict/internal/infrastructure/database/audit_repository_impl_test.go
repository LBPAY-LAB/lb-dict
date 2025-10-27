package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/infrastructure/database"
)

func createAuditTable(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()

	// Create audit schema
	_, err := pool.Exec(ctx, `CREATE SCHEMA IF NOT EXISTS audit`)
	require.NoError(t, err)

	// Create entry_events table
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS audit.entry_events (
			event_id UUID PRIMARY KEY,
			entity_type VARCHAR(50) NOT NULL,
			entity_id UUID NOT NULL,
			event_type VARCHAR(50) NOT NULL,
			user_id UUID,
			old_values JSONB,
			new_values JSONB,
			metadata JSONB,
			ip_address VARCHAR(45),
			user_agent TEXT,
			occurred_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	require.NoError(t, err)

	// Create indexes
	_, err = pool.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_audit_entity ON audit.entry_events(entity_type, entity_id);
		CREATE INDEX IF NOT EXISTS idx_audit_user ON audit.entry_events(user_id);
		CREATE INDEX IF NOT EXISTS idx_audit_occurred ON audit.entry_events(occurred_at);
	`)
	require.NoError(t, err)
}

func TestAuditRepo_Create_Success(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	createAuditTable(t, pool)

	repo := database.NewPostgresAuditRepository(pool)

	userID := uuid.New()
	event := &entities.AuditEvent{
		EventID:    uuid.New(),
		EntityType: entities.EntityTypeEntry,
		EntityID:   uuid.New(),
		EventType:  entities.EventTypeEntryCreated,
		UserID:     &userID,
		OldValues:  map[string]interface{}{},
		NewValues: map[string]interface{}{
			"key_value": "test@example.com",
			"status":    "ACTIVE",
		},
		Metadata: map[string]interface{}{
			"source": "api",
		},
		IPAddress:  "192.168.1.1",
		UserAgent:  "Mozilla/5.0",
		OccurredAt: time.Now(),
	}

	err := repo.Create(context.Background(), event)
	assert.NoError(t, err)

	// Verify event was created
	found, err := repo.FindByID(context.Background(), event.EventID)
	assert.NoError(t, err)
	assert.Equal(t, event.EventID, found.EventID)
	assert.Equal(t, event.EntityType, found.EntityType)
}

func TestAuditRepo_FindByEntity_Success(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	createAuditTable(t, pool)

	repo := database.NewPostgresAuditRepository(pool)

	entityID := uuid.New()

	// Create multiple events for same entity
	for i := 0; i < 3; i++ {
		userID := uuid.New()
		event := &entities.AuditEvent{
			EventID:    uuid.New(),
			EntityType: entities.EntityTypeEntry,
			EntityID:   entityID,
			EventType:  entities.EventTypeEntryUpdated,
			UserID:     &userID,
			OldValues:  map[string]interface{}{"status": "PENDING"},
			NewValues:  map[string]interface{}{"status": "ACTIVE"},
			Metadata:   map[string]interface{}{},
			OccurredAt: time.Now(),
		}
		err := repo.Create(context.Background(), event)
		require.NoError(t, err)
	}

	// Find by entity
	events, err := repo.FindByEntityID(context.Background(), entities.EntityTypeEntry, entityID, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, events, 3)

	for _, event := range events {
		assert.Equal(t, entityID, event.EntityID)
	}
}

func TestAuditRepo_List_TimeRange(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	createAuditTable(t, pool)

	repo := database.NewPostgresAuditRepository(pool)

	now := time.Now()

	// Create events at different times
	userID1 := uuid.New()
	event1 := &entities.AuditEvent{
		EventID:    uuid.New(),
		EntityType: entities.EntityTypeEntry,
		EntityID:   uuid.New(),
		EventType:  entities.EventTypeEntryCreated,
		UserID:     &userID1,
		OldValues:  map[string]interface{}{},
		NewValues:  map[string]interface{}{"status": "ACTIVE"},
		Metadata:   map[string]interface{}{},
		OccurredAt: now.Add(-2 * time.Hour),
	}
	err := repo.Create(context.Background(), event1)
	require.NoError(t, err)

	userID2 := uuid.New()
	event2 := &entities.AuditEvent{
		EventID:    uuid.New(),
		EntityType: entities.EntityTypeEntry,
		EntityID:   uuid.New(),
		EventType:  entities.EventTypeEntryCreated,
		UserID:     &userID2,
		OldValues:  map[string]interface{}{},
		NewValues:  map[string]interface{}{"status": "ACTIVE"},
		Metadata:   map[string]interface{}{},
		OccurredAt: now.Add(-5 * time.Hour),
	}
	err = repo.Create(context.Background(), event2)
	require.NoError(t, err)

	// Find events in time range (last 3 hours)
	from := now.Add(-3 * time.Hour)
	to := now
	events, err := repo.FindByDateRange(context.Background(), from, to, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, event1.EventID, events[0].EventID)
}

func TestAuditRepo_List_ByUser(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	createAuditTable(t, pool)

	repo := database.NewPostgresAuditRepository(pool)

	userID := uuid.New()

	// Create events for specific user
	for i := 0; i < 2; i++ {
		event := &entities.AuditEvent{
			EventID:    uuid.New(),
			EntityType: entities.EntityTypeEntry,
			EntityID:   uuid.New(),
			EventType:  entities.EventTypeEntryCreated,
			UserID:     &userID,
			OldValues:  map[string]interface{}{},
			NewValues:  map[string]interface{}{"status": "ACTIVE"},
			Metadata:   map[string]interface{}{},
			OccurredAt: time.Now(),
		}
		err := repo.Create(context.Background(), event)
		require.NoError(t, err)
	}

	// Create event for different user
	otherUserID := uuid.New()
	otherEvent := &entities.AuditEvent{
		EventID:    uuid.New(),
		EntityType: entities.EntityTypeEntry,
		EntityID:   uuid.New(),
		EventType:  entities.EventTypeEntryCreated,
		UserID:     &otherUserID,
		OldValues:  map[string]interface{}{},
		NewValues:  map[string]interface{}{"status": "ACTIVE"},
		Metadata:   map[string]interface{}{},
		OccurredAt: time.Now(),
	}
	err := repo.Create(context.Background(), otherEvent)
	require.NoError(t, err)

	// Find by user
	events, err := repo.FindByUserID(context.Background(), userID, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, events, 2)

	for _, event := range events {
		assert.Equal(t, userID, event.UserID)
	}
}
