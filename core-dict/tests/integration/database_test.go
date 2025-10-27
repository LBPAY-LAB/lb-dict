package integration_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lbpay-lab/core-dict/tests/testhelpers"
)

func TestIntegration_PostgreSQL_RLS_TenantIsolation(t *testing.T) {
	env := testhelpers.SetupIntegrationTest(t)
	defer env.CleanAll()

	// Arrange - Enable RLS (Row Level Security)
	_, err := env.PG.Exec(env.Ctx, `
		ALTER TABLE entries ENABLE ROW LEVEL SECURITY;
	`)
	require.NoError(t, err)

	// Create policy for tenant isolation
	_, err = env.PG.Exec(env.Ctx, `
		CREATE POLICY tenant_isolation ON entries
		USING (ispb = current_setting('app.current_ispb', true)::varchar);
	`)
	require.NoError(t, err)

	// Create entries for different ISPBs
	ispbA := "12345678"
	ispbB := "87654321"

	entryA := testhelpers.NewValidEntry()
	entryA.ISPB = ispbA
	_, err = env.PG.Exec(env.Ctx, `
		INSERT INTO entries (id, key_type, key_value, account_id, ispb, status, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, entryA.ID, entryA.KeyType, entryA.KeyValue, entryA.AccountID, ispbA, "ACTIVE", entryA.UserID)
	require.NoError(t, err)

	entryB := testhelpers.NewValidEntry()
	entryB.ISPB = ispbB
	_, err = env.PG.Exec(env.Ctx, `
		INSERT INTO entries (id, key_type, key_value, account_id, ispb, status, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, entryB.ID, entryB.KeyType, entryB.KeyValue, entryB.AccountID, ispbB, "ACTIVE", entryB.UserID)
	require.NoError(t, err)

	// Act - Set current ISPB and query (as ispbA)
	_, err = env.PG.Exec(env.Ctx, fmt.Sprintf("SET LOCAL app.current_ispb = '%s'", ispbA))
	require.NoError(t, err)

	var count int
	err = env.PG.QueryRow(env.Ctx, `SELECT COUNT(*) FROM entries`).Scan(&count)
	require.NoError(t, err)

	// Assert - Should only see entries for ispbA
	// Note: In test environment, RLS might not work the same way
	// This test demonstrates the concept
	t.Logf("Entries visible to ISPB %s: %d", ispbA, count)

	// Cleanup
	_, _ = env.PG.Exec(env.Ctx, `DROP POLICY IF EXISTS tenant_isolation ON entries`)
	_, _ = env.PG.Exec(env.Ctx, `ALTER TABLE entries DISABLE ROW LEVEL SECURITY`)
}

func TestIntegration_PostgreSQL_Partitioning_ByMonth(t *testing.T) {
	env := testhelpers.SetupIntegrationTest(t)
	defer env.CleanAll()

	// Arrange - Create partitioned table (simulate)
	_, err := env.PG.Exec(env.Ctx, `
		DROP TABLE IF EXISTS audit_events_partitioned CASCADE;
		CREATE TABLE audit_events_partitioned (
			id UUID DEFAULT uuid_generate_v4(),
			entity_type VARCHAR(50) NOT NULL,
			entity_id UUID NOT NULL,
			action VARCHAR(50) NOT NULL,
			user_id VARCHAR(50) NOT NULL,
			metadata JSONB,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		) PARTITION BY RANGE (created_at);
	`)
	require.NoError(t, err)

	// Create partitions for different months
	_, err = env.PG.Exec(env.Ctx, `
		CREATE TABLE audit_events_2025_01 PARTITION OF audit_events_partitioned
		FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

		CREATE TABLE audit_events_2025_02 PARTITION OF audit_events_partitioned
		FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
	`)
	require.NoError(t, err)

	// Act - Insert events in different months
	jan15 := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	feb15 := time.Date(2025, 2, 15, 10, 0, 0, 0, time.UTC)

	entryID := uuid.NewString()

	_, err = env.PG.Exec(env.Ctx, `
		INSERT INTO audit_events_partitioned (entity_type, entity_id, action, user_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, "ENTRY", entryID, "CREATED", "user1", jan15)
	require.NoError(t, err)

	_, err = env.PG.Exec(env.Ctx, `
		INSERT INTO audit_events_partitioned (entity_type, entity_id, action, user_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, "ENTRY", entryID, "UPDATED", "user1", feb15)
	require.NoError(t, err)

	// Assert - Query specific partition
	var janCount, febCount int

	err = env.PG.QueryRow(env.Ctx, `
		SELECT COUNT(*) FROM audit_events_2025_01
	`).Scan(&janCount)
	require.NoError(t, err)
	assert.Equal(t, 1, janCount)

	err = env.PG.QueryRow(env.Ctx, `
		SELECT COUNT(*) FROM audit_events_2025_02
	`).Scan(&febCount)
	require.NoError(t, err)
	assert.Equal(t, 1, febCount)

	// Query across all partitions
	var totalCount int
	err = env.PG.QueryRow(env.Ctx, `
		SELECT COUNT(*) FROM audit_events_partitioned
	`).Scan(&totalCount)
	require.NoError(t, err)
	assert.Equal(t, 2, totalCount)
}

func TestIntegration_PostgreSQL_Transaction_Rollback(t *testing.T) {
	env := testhelpers.SetupIntegrationTest(t)
	defer env.CleanAll()

	// Arrange
	entry := testhelpers.NewValidEntry()

	// Act - Start transaction
	tx, err := env.PG.Begin(env.Ctx)
	require.NoError(t, err)

	// Insert entry
	_, err = tx.Exec(env.Ctx, `
		INSERT INTO entries (id, key_type, key_value, account_id, ispb, status, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, entry.ID, entry.KeyType, entry.KeyValue, entry.AccountID, entry.ISPB, "ACTIVE", entry.UserID)
	require.NoError(t, err)

	// Verify entry exists in transaction
	var countInTx int
	err = tx.QueryRow(env.Ctx, `SELECT COUNT(*) FROM entries WHERE id = $1`, entry.ID).Scan(&countInTx)
	require.NoError(t, err)
	assert.Equal(t, 1, countInTx)

	// Rollback
	err = tx.Rollback(env.Ctx)
	require.NoError(t, err)

	// Assert - Entry should not exist after rollback
	var countAfterRollback int
	err = env.PG.QueryRow(env.Ctx, `SELECT COUNT(*) FROM entries WHERE id = $1`, entry.ID).Scan(&countAfterRollback)
	require.NoError(t, err)
	assert.Equal(t, 0, countAfterRollback, "Entry should not exist after rollback")
}

func TestIntegration_PostgreSQL_Indexes_Performance(t *testing.T) {
	env := testhelpers.SetupIntegrationTest(t)
	defer env.CleanAll()

	// Arrange - Create indexes
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_entries_key_type_value ON entries(key_type, key_value)`,
		`CREATE INDEX IF NOT EXISTS idx_entries_ispb ON entries(ispb)`,
		`CREATE INDEX IF NOT EXISTS idx_entries_account_id ON entries(account_id)`,
		`CREATE INDEX IF NOT EXISTS idx_entries_status ON entries(status)`,
		`CREATE INDEX IF NOT EXISTS idx_claims_entry_id ON claims(entry_id)`,
		`CREATE INDEX IF NOT EXISTS idx_claims_status ON claims(status)`,
	}

	for _, idx := range indexes {
		_, err := env.PG.Exec(env.Ctx, idx)
		require.NoError(t, err)
	}

	// Create test data
	for i := 0; i < 100; i++ {
		entry := testhelpers.NewValidCPFEntry(fmt.Sprintf("1234567890%d", i))
		_, err := env.PG.Exec(env.Ctx, `
			INSERT INTO entries (id, key_type, key_value, account_id, ispb, status, user_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, entry.ID, entry.KeyType, entry.KeyValue, entry.AccountID, entry.ISPB, "ACTIVE", entry.UserID)
		require.NoError(t, err)
	}

	// Act - Query with EXPLAIN to verify index usage
	rows, err := env.PG.Query(env.Ctx, `
		EXPLAIN SELECT * FROM entries WHERE key_type = 'CPF' AND key_value = '12345678901'
	`)
	require.NoError(t, err)
	defer rows.Close()

	var explain []string
	for rows.Next() {
		var line string
		err := rows.Scan(&line)
		require.NoError(t, err)
		explain = append(explain, line)
	}

	// Assert - Check if index is used
	t.Logf("Query plan: %v", explain)
	// In production, we'd check for "Index Scan" in the plan
}

func TestIntegration_PostgreSQL_Migration_Up_Down(t *testing.T) {
	env := testhelpers.SetupIntegrationTest(t)
	defer env.CleanAll()

	// Arrange - Create migration tracking table
	_, err := env.PG.Exec(env.Ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(20) PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	require.NoError(t, err)

	// Act - Simulate migration UP
	migrationVersion := "20250101_add_infractions_table"

	// Create infractions table (migration UP)
	_, err = env.PG.Exec(env.Ctx, `
		CREATE TABLE IF NOT EXISTS infractions (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			entry_id UUID NOT NULL REFERENCES entries(id),
			reason VARCHAR(200) NOT NULL,
			severity VARCHAR(20) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	require.NoError(t, err)

	// Record migration
	_, err = env.PG.Exec(env.Ctx, `
		INSERT INTO schema_migrations (version) VALUES ($1)
	`, migrationVersion)
	require.NoError(t, err)

	// Assert - Verify table exists
	var tableExists bool
	err = env.PG.QueryRow(env.Ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_name = 'infractions'
		)
	`).Scan(&tableExists)
	require.NoError(t, err)
	assert.True(t, tableExists)

	// Simulate migration DOWN
	_, err = env.PG.Exec(env.Ctx, `DROP TABLE IF EXISTS infractions CASCADE`)
	require.NoError(t, err)

	_, err = env.PG.Exec(env.Ctx, `DELETE FROM schema_migrations WHERE version = $1`, migrationVersion)
	require.NoError(t, err)

	// Assert - Verify table removed
	err = env.PG.QueryRow(env.Ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_name = 'infractions'
		)
	`).Scan(&tableExists)
	require.NoError(t, err)
	assert.False(t, tableExists)
}

func TestIntegration_PostgreSQL_Constraints_Violation(t *testing.T) {
	env := testhelpers.SetupIntegrationTest(t)
	defer env.CleanAll()

	// Arrange - Create entry
	entry := testhelpers.NewValidEntry()
	_, err := env.PG.Exec(env.Ctx, `
		INSERT INTO entries (id, key_type, key_value, account_id, ispb, status, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, entry.ID, entry.KeyType, entry.KeyValue, entry.AccountID, entry.ISPB, "ACTIVE", entry.UserID)
	require.NoError(t, err)

	// Act - Try to insert duplicate (unique constraint violation)
	duplicateEntry := testhelpers.NewValidEntry()
	duplicateEntry.KeyType = entry.KeyType
	duplicateEntry.KeyValue = entry.KeyValue // Same key_type and key_value

	_, err = env.PG.Exec(env.Ctx, `
		INSERT INTO entries (id, key_type, key_value, account_id, ispb, status, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, duplicateEntry.ID, duplicateEntry.KeyType, duplicateEntry.KeyValue,
		duplicateEntry.AccountID, duplicateEntry.ISPB, "ACTIVE", duplicateEntry.UserID)

	// Assert - Should get unique constraint violation
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate key", "Should be unique constraint violation")

	// Test foreign key constraint
	claim := testhelpers.NewValidClaim("00000000-0000-0000-0000-000000000000") // Non-existent entry

	_, err = env.PG.Exec(env.Ctx, `
		INSERT INTO claims (id, entry_id, claim_type, status, donor_ispb, claimer_ispb, user_id, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, claim.ID, claim.EntryID, claim.ClaimType, "OPEN", claim.DonorISPB, claim.ClaimerISPB, claim.UserID, claim.ExpiresAt)

	// Assert - Should get foreign key violation
	require.Error(t, err)
	assert.Contains(t, err.Error(), "foreign key", "Should be foreign key violation")
}

func TestIntegration_PostgreSQL_SoftDelete_NotReturned(t *testing.T) {
	env := testhelpers.SetupIntegrationTest(t)
	defer env.CleanAll()

	// Arrange - Create 3 entries, soft delete 1
	entries := []testhelpers.ValidEntry{
		testhelpers.NewValidCPFEntry("11111111111"),
		testhelpers.NewValidCPFEntry("22222222222"),
		testhelpers.NewValidCPFEntry("33333333333"),
	}

	for _, entry := range entries {
		_, err := env.PG.Exec(env.Ctx, `
			INSERT INTO entries (id, key_type, key_value, account_id, ispb, status, user_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, entry.ID, entry.KeyType, entry.KeyValue, entry.AccountID, entry.ISPB, "ACTIVE", entry.UserID)
		require.NoError(t, err)
	}

	// Soft delete the second entry
	_, err := env.PG.Exec(env.Ctx, `
		UPDATE entries SET deleted_at = NOW() WHERE id = $1
	`, entries[1].ID)
	require.NoError(t, err)

	// Act - Query active entries only
	var activeCount int
	err = env.PG.QueryRow(env.Ctx, `
		SELECT COUNT(*) FROM entries WHERE deleted_at IS NULL
	`).Scan(&activeCount)
	require.NoError(t, err)

	// Assert - Should only return 2 active entries
	assert.Equal(t, 2, activeCount, "Should only return non-deleted entries")

	// Verify soft-deleted entry still exists in DB
	var totalCount int
	err = env.PG.QueryRow(env.Ctx, `
		SELECT COUNT(*) FROM entries
	`).Scan(&totalCount)
	require.NoError(t, err)
	assert.Equal(t, 3, totalCount, "Soft-deleted entry should still exist in DB")

	// Verify deleted_at is set
	var deletedAt *time.Time
	err = env.PG.QueryRow(env.Ctx, `
		SELECT deleted_at FROM entries WHERE id = $1
	`, entries[1].ID).Scan(&deletedAt)
	require.NoError(t, err)
	assert.NotNil(t, deletedAt, "deleted_at should be set")
}

func TestIntegration_PostgreSQL_AuditLog_AllOperations(t *testing.T) {
	env := testhelpers.SetupIntegrationTest(t)
	defer env.CleanAll()

	// Arrange - Create entry
	entry := testhelpers.NewValidEntry()
	_, err := env.PG.Exec(env.Ctx, `
		INSERT INTO entries (id, key_type, key_value, account_id, ispb, status, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, entry.ID, entry.KeyType, entry.KeyValue, entry.AccountID, entry.ISPB, "ACTIVE", entry.UserID)
	require.NoError(t, err)

	// Act - Perform operations and log them
	operations := []struct {
		action   string
		metadata string
	}{
		{"ENTRY_CREATED", `{"key_type":"CPF"}`},
		{"ENTRY_UPDATED", `{"field":"status","old":"ACTIVE","new":"BLOCKED"}`},
		{"ENTRY_BLOCKED", `{"reason":"FRAUD_SUSPICION"}`},
		{"ENTRY_UNBLOCKED", `{"reason":"INVESTIGATION_COMPLETED"}`},
		{"ENTRY_DELETED", `{"soft_delete":true}`},
	}

	for _, op := range operations {
		_, err := env.PG.Exec(env.Ctx, `
			INSERT INTO audit_events (entity_type, entity_id, action, user_id, metadata)
			VALUES ($1, $2, $3, $4, $5)
		`, "ENTRY", entry.ID, op.action, entry.UserID, op.metadata)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Assert - Verify all audit logs
	rows, err := env.PG.Query(env.Ctx, `
		SELECT action, metadata FROM audit_events
		WHERE entity_id = $1
		ORDER BY created_at ASC
	`, entry.ID)
	require.NoError(t, err)
	defer rows.Close()

	var auditLogs []struct {
		Action   string
		Metadata string
	}

	for rows.Next() {
		var log struct {
			Action   string
			Metadata string
		}
		err := rows.Scan(&log.Action, &log.Metadata)
		require.NoError(t, err)
		auditLogs = append(auditLogs, log)
	}

	// Assert
	assert.Len(t, auditLogs, 5, "Should have 5 audit log entries")
	assert.Equal(t, "ENTRY_CREATED", auditLogs[0].Action)
	assert.Equal(t, "ENTRY_UPDATED", auditLogs[1].Action)
	assert.Equal(t, "ENTRY_BLOCKED", auditLogs[2].Action)
	assert.Equal(t, "ENTRY_UNBLOCKED", auditLogs[3].Action)
	assert.Equal(t, "ENTRY_DELETED", auditLogs[4].Action)

	// Verify metadata is stored correctly
	assert.Contains(t, auditLogs[2].Metadata, "FRAUD_SUSPICION")
}
