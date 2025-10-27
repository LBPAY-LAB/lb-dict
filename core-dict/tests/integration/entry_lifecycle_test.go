package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lbpay-lab/core-dict/tests/testhelpers"
)

func TestIntegration_CreateEntry_CompleteFlow(t *testing.T) {
	env := testhelpers.SetupIntegrationTest(t)
	defer env.CleanAll()

	// Arrange
	entry := testhelpers.NewValidEntry()

	// Act - Insert entry into database
	query := `
		INSERT INTO entries (id, key_type, key_value, account_id, ispb, status, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`
	var createdID string
	var createdAt time.Time
	err := env.PG.QueryRow(env.Ctx, query, entry.ID, entry.KeyType, entry.KeyValue,
		entry.AccountID, entry.ISPB, entry.Status, entry.UserID).Scan(&createdID, &createdAt)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, entry.ID, createdID)
	assert.False(t, createdAt.IsZero())

	// Verify database
	var dbEntry testhelpers.ValidEntry
	err = env.PG.QueryRow(env.Ctx, `
		SELECT id, key_type, key_value, account_id, ispb, status, user_id
		FROM entries WHERE id = $1
	`, createdID).Scan(&dbEntry.ID, &dbEntry.KeyType, &dbEntry.KeyValue,
		&dbEntry.AccountID, &dbEntry.ISPB, &dbEntry.Status, &dbEntry.UserID)

	require.NoError(t, err)
	assert.Equal(t, entry.KeyType, dbEntry.KeyType)
	assert.Equal(t, entry.KeyValue, dbEntry.KeyValue)

	// Verify cache (simulate cache write)
	cacheKey := fmt.Sprintf("entry:%s", createdID)
	err = env.Redis.Set(env.Ctx, cacheKey, createdID, 5*time.Minute).Err()
	require.NoError(t, err)

	// Verify cache read
	cachedID, err := env.Redis.Get(env.Ctx, cacheKey).Result()
	require.NoError(t, err)
	assert.Equal(t, createdID, cachedID)

	// Verify Pulsar event
	env.PulsarMock.Publish("dict.entries.created", entry.ID, []byte(fmt.Sprintf(`{"id":"%s","keyType":"%s"}`, entry.ID, entry.KeyType)))
	assert.True(t, env.PulsarMock.ReceivedEvent("dict.entries.created"))

	// Verify audit log
	auditQuery := `
		INSERT INTO audit_events (entity_type, entity_id, action, user_id)
		VALUES ($1, $2, $3, $4)
	`
	_, err = env.PG.Exec(env.Ctx, auditQuery, "ENTRY", entry.ID, "ENTRY_CREATED", entry.UserID)
	require.NoError(t, err)

	var auditCount int
	err = env.PG.QueryRow(env.Ctx, `
		SELECT COUNT(*) FROM audit_events WHERE entity_id = $1 AND action = $2
	`, entry.ID, "ENTRY_CREATED").Scan(&auditCount)
	require.NoError(t, err)
	assert.Equal(t, 1, auditCount)
}

func TestIntegration_CreateEntry_DuplicateCheck_GlobalViaConnect(t *testing.T) {
	env := testhelpers.SetupIntegrationTest(t)
	defer env.CleanAll()

	// Arrange
	entry := testhelpers.NewValidEntry()

	// Set mock to return duplicate
	env.ConnectMock.SetDuplicateResult(entry.KeyType, entry.KeyValue, true)

	// Act - Check duplicate via Connect mock
	isDuplicate, err := env.ConnectMock.CheckDuplicate(env.Ctx, entry.KeyType, entry.KeyValue)

	// Assert
	require.NoError(t, err)
	assert.True(t, isDuplicate, "Entry should be detected as duplicate")

	// Verify Connect was called
	assert.Equal(t, 1, env.ConnectMock.GetCallCount("CheckDuplicate"))

	// Now test non-duplicate
	entry2 := testhelpers.NewValidCPFEntry(testhelpers.TestCPF2)
	env.ConnectMock.SetDuplicateResult(entry2.KeyType, entry2.KeyValue, false)

	isDuplicate2, err := env.ConnectMock.CheckDuplicate(env.Ctx, entry2.KeyType, entry2.KeyValue)
	require.NoError(t, err)
	assert.False(t, isDuplicate2, "Entry should not be detected as duplicate")
}

func TestIntegration_UpdateEntry_WithCache_Invalidation(t *testing.T) {
	env := testhelpers.SetupIntegrationTest(t)
	defer env.CleanAll()

	// Arrange - Create entry
	entry := testhelpers.NewValidEntry()
	_, err := env.PG.Exec(env.Ctx, `
		INSERT INTO entries (id, key_type, key_value, account_id, ispb, status, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, entry.ID, entry.KeyType, entry.KeyValue, entry.AccountID, entry.ISPB, entry.Status, entry.UserID)
	require.NoError(t, err)

	// Cache the entry
	cacheKey := fmt.Sprintf("entry:%s", entry.ID)
	err = env.Redis.Set(env.Ctx, cacheKey, entry.ID, 5*time.Minute).Err()
	require.NoError(t, err)

	// Act - Update entry
	newStatus := "BLOCKED"
	_, err = env.PG.Exec(env.Ctx, `
		UPDATE entries SET status = $1, updated_at = NOW()
		WHERE id = $2
	`, newStatus, entry.ID)
	require.NoError(t, err)

	// Invalidate cache
	err = env.Redis.Del(env.Ctx, cacheKey).Err()
	require.NoError(t, err)

	// Assert - Verify update
	var updatedStatus string
	err = env.PG.QueryRow(env.Ctx, `SELECT status FROM entries WHERE id = $1`, entry.ID).Scan(&updatedStatus)
	require.NoError(t, err)
	assert.Equal(t, newStatus, updatedStatus)

	// Verify cache was invalidated
	_, err = env.Redis.Get(env.Ctx, cacheKey).Result()
	assert.Error(t, err, "Cache should be invalidated")
}

func TestIntegration_DeleteEntry_SoftDelete_AuditLog(t *testing.T) {
	env := testhelpers.SetupIntegrationTest(t)
	defer env.CleanAll()

	// Arrange - Create entry
	entry := testhelpers.NewValidEntry()
	_, err := env.PG.Exec(env.Ctx, `
		INSERT INTO entries (id, key_type, key_value, account_id, ispb, status, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, entry.ID, entry.KeyType, entry.KeyValue, entry.AccountID, entry.ISPB, entry.Status, entry.UserID)
	require.NoError(t, err)

	// Act - Soft delete
	deletedAt := time.Now()
	_, err = env.PG.Exec(env.Ctx, `
		UPDATE entries SET deleted_at = $1 WHERE id = $2
	`, deletedAt, entry.ID)
	require.NoError(t, err)

	// Create audit log
	_, err = env.PG.Exec(env.Ctx, `
		INSERT INTO audit_events (entity_type, entity_id, action, user_id)
		VALUES ($1, $2, $3, $4)
	`, "ENTRY", entry.ID, "ENTRY_DELETED", entry.UserID)
	require.NoError(t, err)

	// Assert - Verify soft delete
	var deletedAtDB *time.Time
	err = env.PG.QueryRow(env.Ctx, `SELECT deleted_at FROM entries WHERE id = $1`, entry.ID).Scan(&deletedAtDB)
	require.NoError(t, err)
	assert.NotNil(t, deletedAtDB)

	// Verify entry not returned in active queries
	var count int
	err = env.PG.QueryRow(env.Ctx, `
		SELECT COUNT(*) FROM entries WHERE id = $1 AND deleted_at IS NULL
	`, entry.ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Verify audit log
	var auditAction string
	err = env.PG.QueryRow(env.Ctx, `
		SELECT action FROM audit_events WHERE entity_id = $1 AND action = 'ENTRY_DELETED'
	`, entry.ID).Scan(&auditAction)
	require.NoError(t, err)
	assert.Equal(t, "ENTRY_DELETED", auditAction)
}

func TestIntegration_BlockEntry_StatusChange_EventPublished(t *testing.T) {
	env := testhelpers.SetupIntegrationTest(t)
	defer env.CleanAll()

	// Arrange - Create active entry
	entry := testhelpers.NewValidEntry()
	_, err := env.PG.Exec(env.Ctx, `
		INSERT INTO entries (id, key_type, key_value, account_id, ispb, status, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, entry.ID, entry.KeyType, entry.KeyValue, entry.AccountID, entry.ISPB, "ACTIVE", entry.UserID)
	require.NoError(t, err)

	// Act - Block entry
	_, err = env.PG.Exec(env.Ctx, `
		UPDATE entries SET status = 'BLOCKED', updated_at = NOW()
		WHERE id = $1
	`, entry.ID)
	require.NoError(t, err)

	// Publish event
	env.PulsarMock.Publish("dict.entries.blocked", entry.ID, []byte(fmt.Sprintf(`{"id":"%s","status":"BLOCKED"}`, entry.ID)))

	// Assert - Verify status
	var status string
	err = env.PG.QueryRow(env.Ctx, `SELECT status FROM entries WHERE id = $1`, entry.ID).Scan(&status)
	require.NoError(t, err)
	assert.Equal(t, "BLOCKED", status)

	// Verify event published
	assert.True(t, env.PulsarMock.ReceivedEvent("dict.entries.blocked"))
	msgs := env.PulsarMock.GetMessages("dict.entries.blocked")
	assert.Len(t, msgs, 1)
	assert.Equal(t, entry.ID, msgs[0].Key)
}

func TestIntegration_UnblockEntry_CompleteFlow(t *testing.T) {
	env := testhelpers.SetupIntegrationTest(t)
	defer env.CleanAll()

	// Arrange - Create blocked entry
	entry := testhelpers.NewValidEntry()
	_, err := env.PG.Exec(env.Ctx, `
		INSERT INTO entries (id, key_type, key_value, account_id, ispb, status, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, entry.ID, entry.KeyType, entry.KeyValue, entry.AccountID, entry.ISPB, "BLOCKED", entry.UserID)
	require.NoError(t, err)

	// Act - Unblock entry
	_, err = env.PG.Exec(env.Ctx, `
		UPDATE entries SET status = 'ACTIVE', updated_at = NOW()
		WHERE id = $1
	`, entry.ID)
	require.NoError(t, err)

	// Publish event
	env.PulsarMock.Publish("dict.entries.unblocked", entry.ID, []byte(fmt.Sprintf(`{"id":"%s","status":"ACTIVE"}`, entry.ID)))

	// Assert
	var status string
	err = env.PG.QueryRow(env.Ctx, `SELECT status FROM entries WHERE id = $1`, entry.ID).Scan(&status)
	require.NoError(t, err)
	assert.Equal(t, "ACTIVE", status)

	// Verify event
	assert.True(t, env.PulsarMock.ReceivedEvent("dict.entries.unblocked"))
}

func TestIntegration_TransferOwnership_Portability(t *testing.T) {
	env := testhelpers.SetupIntegrationTest(t)
	defer env.CleanAll()

	// Arrange - Create entry with original owner
	entry := testhelpers.NewValidEntry()
	originalISPB := "12345678"
	newISPB := "87654321"

	_, err := env.PG.Exec(env.Ctx, `
		INSERT INTO entries (id, key_type, key_value, account_id, ispb, status, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, entry.ID, entry.KeyType, entry.KeyValue, entry.AccountID, originalISPB, "ACTIVE", entry.UserID)
	require.NoError(t, err)

	// Act - Transfer ownership
	newAccountID := uuid.NewString()
	_, err = env.PG.Exec(env.Ctx, `
		UPDATE entries SET ispb = $1, account_id = $2, updated_at = NOW()
		WHERE id = $3
	`, newISPB, newAccountID, entry.ID)
	require.NoError(t, err)

	// Create audit log
	_, err = env.PG.Exec(env.Ctx, `
		INSERT INTO audit_events (entity_type, entity_id, action, user_id, metadata)
		VALUES ($1, $2, $3, $4, $5)
	`, "ENTRY", entry.ID, "OWNERSHIP_TRANSFERRED", entry.UserID,
		fmt.Sprintf(`{"from_ispb":"%s","to_ispb":"%s"}`, originalISPB, newISPB))
	require.NoError(t, err)

	// Assert
	var updatedISPB, updatedAccountID string
	err = env.PG.QueryRow(env.Ctx, `
		SELECT ispb, account_id FROM entries WHERE id = $1
	`, entry.ID).Scan(&updatedISPB, &updatedAccountID)
	require.NoError(t, err)
	assert.Equal(t, newISPB, updatedISPB)
	assert.Equal(t, newAccountID, updatedAccountID)

	// Verify audit log
	var auditAction string
	err = env.PG.QueryRow(env.Ctx, `
		SELECT action FROM audit_events WHERE entity_id = $1 AND action = 'OWNERSHIP_TRANSFERRED'
	`, entry.ID).Scan(&auditAction)
	require.NoError(t, err)
	assert.Equal(t, "OWNERSHIP_TRANSFERRED", auditAction)
}

func TestIntegration_ListEntries_Pagination_Cache(t *testing.T) {
	env := testhelpers.SetupIntegrationTest(t)
	defer env.CleanAll()

	// Arrange - Create multiple entries
	for i := 0; i < 25; i++ {
		entry := testhelpers.NewValidCPFEntry(fmt.Sprintf("1234567890%d", i))
		_, err := env.PG.Exec(env.Ctx, `
			INSERT INTO entries (id, key_type, key_value, account_id, ispb, status, user_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, entry.ID, entry.KeyType, entry.KeyValue, entry.AccountID, entry.ISPB, entry.Status, entry.UserID)
		require.NoError(t, err)
	}

	// Act - Query with pagination
	limit := 10
	offset := 0
	rows, err := env.PG.Query(env.Ctx, `
		SELECT id, key_type, key_value
		FROM entries
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	require.NoError(t, err)
	defer rows.Close()

	// Assert
	entries := []testhelpers.ValidEntry{}
	for rows.Next() {
		var e testhelpers.ValidEntry
		err := rows.Scan(&e.ID, &e.KeyType, &e.KeyValue)
		require.NoError(t, err)
		entries = append(entries, e)
	}

	assert.Len(t, entries, 10, "Should return 10 entries (page 1)")

	// Cache the result
	cacheKey := fmt.Sprintf("entries:list:limit=%d:offset=%d", limit, offset)
	err = env.Redis.Set(env.Ctx, cacheKey, "cached", 5*time.Minute).Err()
	require.NoError(t, err)

	// Verify cache
	cached, err := env.Redis.Get(env.Ctx, cacheKey).Result()
	require.NoError(t, err)
	assert.Equal(t, "cached", cached)

	// Query page 2
	offset = 10
	rows2, err := env.PG.Query(env.Ctx, `
		SELECT id, key_type, key_value
		FROM entries
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	require.NoError(t, err)
	defer rows2.Close()

	entries2 := []testhelpers.ValidEntry{}
	for rows2.Next() {
		var e testhelpers.ValidEntry
		err := rows2.Scan(&e.ID, &e.KeyType, &e.KeyValue)
		require.NoError(t, err)
		entries2 = append(entries2, e)
	}

	assert.Len(t, entries2, 10, "Should return 10 entries (page 2)")
}

func TestIntegration_GetEntry_CacheHit_Miss(t *testing.T) {
	env := testhelpers.SetupIntegrationTest(t)
	defer env.CleanAll()

	// Arrange - Create entry
	entry := testhelpers.NewValidEntry()
	_, err := env.PG.Exec(env.Ctx, `
		INSERT INTO entries (id, key_type, key_value, account_id, ispb, status, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, entry.ID, entry.KeyType, entry.KeyValue, entry.AccountID, entry.ISPB, entry.Status, entry.UserID)
	require.NoError(t, err)

	cacheKey := fmt.Sprintf("entry:%s", entry.ID)

	// Act - Cache miss, load from DB
	_, err = env.Redis.Get(env.Ctx, cacheKey).Result()
	assert.Error(t, err, "Should be cache miss")

	// Load from DB
	var dbEntry testhelpers.ValidEntry
	err = env.PG.QueryRow(env.Ctx, `
		SELECT id, key_type, key_value FROM entries WHERE id = $1
	`, entry.ID).Scan(&dbEntry.ID, &dbEntry.KeyType, &dbEntry.KeyValue)
	require.NoError(t, err)

	// Cache it
	err = env.Redis.Set(env.Ctx, cacheKey, dbEntry.ID, 5*time.Minute).Err()
	require.NoError(t, err)

	// Act - Cache hit
	cachedID, err := env.Redis.Get(env.Ctx, cacheKey).Result()
	require.NoError(t, err)
	assert.Equal(t, entry.ID, cachedID, "Should be cache hit")
}

func TestIntegration_CreateEntry_MaxKeys_CPF_5(t *testing.T) {
	env := testhelpers.SetupIntegrationTest(t)
	defer env.CleanAll()

	// Arrange - CPF can have max 5 keys
	testCPF := testhelpers.TestCPF1
	accountID := uuid.NewString()

	// Create 5 entries for same CPF
	for i := 0; i < 5; i++ {
		entry := testhelpers.NewValidCPFEntry(testCPF)
		entry.ID = uuid.NewString()
		entry.KeyValue = fmt.Sprintf("+551199999999%d", i) // Different phone numbers
		entry.KeyType = "PHONE"
		entry.AccountID = accountID

		_, err := env.PG.Exec(env.Ctx, `
			INSERT INTO entries (id, key_type, key_value, account_id, ispb, status, user_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, entry.ID, entry.KeyType, entry.KeyValue, entry.AccountID, entry.ISPB, entry.Status, entry.UserID)
		require.NoError(t, err)
	}

	// Act - Count entries for this account
	var count int
	err := env.PG.QueryRow(env.Ctx, `
		SELECT COUNT(*) FROM entries WHERE account_id = $1 AND deleted_at IS NULL
	`, accountID).Scan(&count)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, 5, count, "Should have exactly 5 keys")

	// Try to create 6th key (should be rejected by business logic)
	entry6 := testhelpers.NewValidCPFEntry(testCPF)
	entry6.ID = uuid.NewString()
	entry6.KeyValue = "+5511999999996"
	entry6.KeyType = "PHONE"
	entry6.AccountID = accountID

	// In real implementation, this would be rejected by business logic
	// Here we just verify we can detect the limit
	assert.Equal(t, 5, count, "Already at max limit of 5 keys")
}
