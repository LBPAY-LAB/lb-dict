package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	bridgev1 "github.com/lbpay-lab/dict-contracts/gen/proto/bridge/v1"
	commonv1 "github.com/lbpay-lab/dict-contracts/gen/proto/common/v1"
	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// VSyncActivities contains all Temporal activities for VSYNC operations
type VSyncActivities struct {
	logger       *logrus.Logger
	entryRepo    *repositories.EntryRepository
	reportRepo   *repositories.SyncReportRepository
	bridgeClient BridgeClient // Interface for Bridge gRPC client
	// Add when implementing Pulsar notifications:
	// pulsarProducer *pulsar.Producer
}

// BridgeClient interface for calling Bridge service (for testing/mocking)
type BridgeClient interface {
	SearchEntries(ctx context.Context, req *bridgev1.SearchEntriesRequest) (*bridgev1.SearchEntriesResponse, error)
}

// NewVSyncActivities creates a new instance of VSyncActivities
func NewVSyncActivities(
	logger *logrus.Logger,
	entryRepo *repositories.EntryRepository,
	reportRepo *repositories.SyncReportRepository,
	bridgeClient BridgeClient,
) *VSyncActivities {
	return &VSyncActivities{
		logger:       logger,
		entryRepo:    entryRepo,
		reportRepo:   reportRepo,
		bridgeClient: bridgeClient,
	}
}

// BacenEntry represents a DICT entry from Bacen's registry
type BacenEntry struct {
	Key             string    `json:"key"`               // PIX key
	KeyType         string    `json:"key_type"`          // CPF, CNPJ, EMAIL, PHONE, EVP
	ParticipantISPB string    `json:"participant_ispb"`  // ISPB code
	AccountBranch   string    `json:"account_branch"`    // Branch number
	AccountNumber   string    `json:"account_number"`    // Account number
	AccountType     string    `json:"account_type"`      // CACC, SLRY, SVGS, TRAN
	OwnerType       string    `json:"owner_type"`        // NATURAL_PERSON, LEGAL_PERSON
	OwnerName       string    `json:"owner_name"`        // Owner full name
	OwnerTaxID      string    `json:"owner_tax_id"`      // CPF or CNPJ
	Status          string    `json:"status"`            // ACTIVE, INACTIVE, BLOCKED
	CreatedAt       time.Time `json:"created_at"`        // Creation timestamp
	UpdatedAt       time.Time `json:"updated_at"`        // Last update timestamp
}

// EntryDiscrepancy represents a difference between Bacen and local database
type EntryDiscrepancy struct {
	Type        string           `json:"type"`         // Type of discrepancy
	Key         string           `json:"key"`          // PIX key
	EntryID     string           `json:"entry_id"`     // Local entry ID (if exists)
	Reason      string           `json:"reason"`       // Human-readable reason
	CreateInput CreateEntryInput `json:"create_input"` // For creating missing entries
	UpdateInput UpdateEntryInput `json:"update_input"` // For updating outdated entries
	BacenData   *BacenEntry      `json:"bacen_data"`   // Data from Bacen
	LocalData   interface{}      `json:"local_data"`   // Data from local DB
}

// Discrepancy types
const (
	DiscrepancyTypeMissingLocal   = "MISSING_LOCAL"   // Entry exists in Bacen but not locally
	DiscrepancyTypeOutdatedLocal  = "OUTDATED_LOCAL"  // Entry exists but data differs
	DiscrepancyTypeMissingBacen   = "MISSING_BACEN"   // Entry exists locally but not in Bacen
)

// FetchBacenEntriesActivity fetches entries from Bacen DICT API
//
// This activity calls the Bacen DICT API to retrieve all entries for a given ISPB.
// For incremental sync, it only fetches entries updated since lastSyncDate.
//
// Parameters:
// - ispb: Participant ISPB (empty = all participants)
// - syncType: "FULL" or "INCREMENTAL"
// - lastSyncDate: For incremental sync, fetch entries updated after this date
//
// Returns:
// - []BacenEntry: List of entries from Bacen DICT
//
// Implementation TODO:
// 1. Initialize Bacen DICT API client (mTLS, ICP-Brasil cert)
// 2. Call appropriate endpoint:
//    - FULL: GET /dict/api/v1/entries?ispb={ispb}
//    - INCREMENTAL: GET /dict/api/v1/entries?ispb={ispb}&updated_after={lastSyncDate}
// 3. Handle pagination (Bacen may return max 1000 entries per page)
// 4. Parse XML/JSON response to BacenEntry structs
// 5. Handle errors (network, auth, rate limit)
func (a *VSyncActivities) FetchBacenEntriesActivity(
	ctx context.Context,
	ispb string,
	syncType string,
	lastSyncDate *time.Time,
) ([]BacenEntry, error) {
	a.logger.WithFields(logrus.Fields{
		"ispb":           ispb,
		"sync_type":      syncType,
		"last_sync_date": lastSyncDate,
	}).Info("Fetching entries from Bacen DICT via Bridge")

	var allEntries []BacenEntry
	pageToken := ""
	pageSize := int32(1000) // Bacen's max page size
	totalFetched := 0

	for {
		// Build SearchEntries request
		req := &bridgev1.SearchEntriesRequest{
			Ispb:      &ispb, // Filter by participant ISPB
			PageSize:  pageSize,
			PageToken: pageToken,
			RequestId: uuid.New().String(),
		}

		a.logger.WithFields(logrus.Fields{
			"ispb":       ispb,
			"page_size":  pageSize,
			"page_token": pageToken,
		}).Debug("Calling Bridge SearchEntries")

		// Call Bridge gRPC
		resp, err := a.bridgeClient.SearchEntries(ctx, req)
		if err != nil {
			// Handle different error types
			if st, ok := status.FromError(err); ok {
				switch st.Code() {
				case codes.Unavailable:
					return nil, fmt.Errorf("bridge service unavailable (check if conn-bridge is running): %w", err)
				case codes.Unauthenticated:
					return nil, fmt.Errorf("bridge authentication failed (mTLS cert issue): %w", err)
				case codes.DeadlineExceeded:
					return nil, fmt.Errorf("bridge request timeout (Bacen may be slow): %w", err)
				case codes.PermissionDenied:
					return nil, fmt.Errorf("bridge permission denied (check ISPB authorization): %w", err)
				default:
					return nil, fmt.Errorf("bridge SearchEntries failed: %w", err)
				}
			}
			return nil, fmt.Errorf("bridge SearchEntries failed: %w", err)
		}

		// Convert proto Entry to BacenEntry
		for _, protoEntry := range resp.Entries {
			bacenEntry := convertProtoEntryToBacenEntry(protoEntry)
			allEntries = append(allEntries, bacenEntry)
		}

		totalFetched += len(resp.Entries)

		a.logger.WithFields(logrus.Fields{
			"page_fetched":  len(resp.Entries),
			"total_fetched": totalFetched,
			"has_more":      resp.NextPageToken != "",
		}).Info("Fetched page from Bacen")

		// Check if there are more pages
		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	a.logger.WithFields(logrus.Fields{
		"ispb":          ispb,
		"total_entries": len(allEntries),
	}).Info("Successfully fetched all entries from Bacen DICT")

	return allEntries, nil
}

// convertProtoEntryToBacenEntry converts Bridge proto Entry to BacenEntry struct
func convertProtoEntryToBacenEntry(protoEntry *bridgev1.Entry) BacenEntry {
	var keyType string
	switch protoEntry.KeyType {
	case commonv1.KeyType_KEY_TYPE_CPF:
		keyType = "CPF"
	case commonv1.KeyType_KEY_TYPE_CNPJ:
		keyType = "CNPJ"
	case commonv1.KeyType_KEY_TYPE_EMAIL:
		keyType = "EMAIL"
	case commonv1.KeyType_KEY_TYPE_PHONE:
		keyType = "PHONE"
	case commonv1.KeyType_KEY_TYPE_EVP:
		keyType = "EVP"
	default:
		keyType = "UNKNOWN"
	}

	var accountType string
	if protoEntry.Account != nil {
		switch protoEntry.Account.AccountType {
		case commonv1.AccountType_ACCOUNT_TYPE_CHECKING:
			accountType = "CACC"
		case commonv1.AccountType_ACCOUNT_TYPE_SAVINGS:
			accountType = "SVGS"
		case commonv1.AccountType_ACCOUNT_TYPE_PAYMENT:
			accountType = "TRAN"
		case commonv1.AccountType_ACCOUNT_TYPE_SALARY:
			accountType = "SLRY"
		default:
			accountType = "UNKNOWN"
		}
	}

	var status string
	switch protoEntry.Status {
	case commonv1.EntryStatus_ENTRY_STATUS_ACTIVE:
		status = "ACTIVE"
	case commonv1.EntryStatus_ENTRY_STATUS_PORTABILITY_PENDING:
		status = "PORTABILITY_PENDING"
	case commonv1.EntryStatus_ENTRY_STATUS_PORTABILITY_CONFIRMED:
		status = "PORTABILITY_CONFIRMED"
	case commonv1.EntryStatus_ENTRY_STATUS_CLAIM_PENDING:
		status = "CLAIM_PENDING"
	case commonv1.EntryStatus_ENTRY_STATUS_DELETED:
		status = "DELETED"
	default:
		status = "UNKNOWN"
	}

	var ownerType string
	if protoEntry.Account != nil {
		switch protoEntry.Account.DocumentType {
		case commonv1.DocumentType_DOCUMENT_TYPE_CPF:
			ownerType = "NATURAL_PERSON"
		case commonv1.DocumentType_DOCUMENT_TYPE_CNPJ:
			ownerType = "LEGAL_PERSON"
		default:
			ownerType = "UNKNOWN"
		}
	}

	entry := BacenEntry{
		Key:             protoEntry.KeyValue,
		KeyType:         keyType,
		ParticipantISPB: "",
		AccountBranch:   "",
		AccountNumber:   "",
		AccountType:     accountType,
		OwnerType:       ownerType,
		OwnerName:       "",
		OwnerTaxID:      "",
		Status:          status,
		CreatedAt:       protoEntry.CreatedAt.AsTime(),
		UpdatedAt:       protoEntry.UpdatedAt.AsTime(),
	}

	// Fill account details if present
	if protoEntry.Account != nil {
		entry.ParticipantISPB = protoEntry.Account.Ispb
		entry.AccountBranch = protoEntry.Account.BranchCode
		entry.AccountNumber = protoEntry.Account.AccountNumber
		entry.OwnerName = protoEntry.Account.AccountHolderName
		entry.OwnerTaxID = protoEntry.Account.AccountHolderDocument
	}

	return entry
}

// CompareEntriesActivity compares Bacen entries with local database
//
// This activity performs a three-way comparison to detect discrepancies:
// 1. Entries in Bacen but not locally (MISSING_LOCAL)
// 2. Entries with different data (OUTDATED_LOCAL)
// 3. Entries locally but not in Bacen (MISSING_BACEN)
//
// Parameters:
// - bacenEntries: List of entries from Bacen DICT
// - ispb: Participant ISPB to filter local entries (empty = all)
//
// Returns:
// - []EntryDiscrepancy: List of detected discrepancies
//
// Implementation TODO:
// 1. Query local database for all entries matching ISPB
// 2. Create hash maps for efficient lookup (key → entry)
// 3. Compare Bacen vs Local:
//    - For each Bacen entry, check if exists locally
//    - If not found: MISSING_LOCAL discrepancy
//    - If found but different: OUTDATED_LOCAL discrepancy
// 4. Compare Local vs Bacen:
//    - For each local entry, check if exists in Bacen
//    - If not found: MISSING_BACEN discrepancy
// 5. Return list of all discrepancies with metadata
func (a *VSyncActivities) CompareEntriesActivity(
	ctx context.Context,
	bacenEntries []BacenEntry,
	ispb string,
) ([]EntryDiscrepancy, error) {
	a.logger.WithFields(logrus.Fields{
		"bacen_entries": len(bacenEntries),
		"ispb":          ispb,
	}).Info("Comparing entries with local database")

	var discrepancies []EntryDiscrepancy

	// Build hash map of Bacen entries for O(1) lookups
	bacenMap := make(map[string]*BacenEntry)
	for i := range bacenEntries {
		bacenMap[bacenEntries[i].Key] = &bacenEntries[i]
	}

	// Query local entries for the participant
	limit := 10000 // High limit for full sync
	localEntries, err := a.entryRepo.ListByParticipant(ctx, ispb, limit, 0)
	if err != nil {
		a.logger.WithError(err).Error("Failed to query local entries")
		return nil, fmt.Errorf("failed to query local entries: %w", err)
	}

	a.logger.WithField("local_entries", len(localEntries)).Info("Local entries queried")

	// Build hash map of local entries
	localMap := make(map[string]*entities.Entry)
	for i := range localEntries {
		localMap[localEntries[i].Key] = localEntries[i]
	}

	// Find MISSING_LOCAL and OUTDATED_LOCAL
	for key, bacenEntry := range bacenMap {
		localEntry, exists := localMap[key]

		if !exists {
			// Entry exists in Bacen but not locally - MISSING_LOCAL
			discrepancies = append(discrepancies, EntryDiscrepancy{
				Type:    DiscrepancyTypeMissingLocal,
				Key:     key,
				Reason:  "Entry exists in Bacen DICT but not in local database",
				BacenData: bacenEntry,
				CreateInput: CreateEntryInput{
					EntryID:         fmt.Sprintf("ENTRY-%s", uuid.New().String()[:8]),
					Key:             bacenEntry.Key,
					KeyType:         bacenEntry.KeyType,
					ParticipantISPB: bacenEntry.ParticipantISPB,
					AccountBranch:   bacenEntry.AccountBranch,
					AccountNumber:   bacenEntry.AccountNumber,
					AccountType:     bacenEntry.AccountType,
					OwnerType:       bacenEntry.OwnerType,
					OwnerName:       bacenEntry.OwnerName,
					OwnerTaxID:      bacenEntry.OwnerTaxID,
				},
			})
		} else {
			// Entry exists in both - check if data differs
			if isDataDifferent(bacenEntry, localEntry) {
				discrepancies = append(discrepancies, EntryDiscrepancy{
					Type:    DiscrepancyTypeOutdatedLocal,
					Key:     key,
					EntryID: localEntry.EntryID,
					Reason:  "Entry data differs between Bacen and local database",
					BacenData: bacenEntry,
					LocalData: localEntry,
					UpdateInput: UpdateEntryInput{
						AccountBranch: &bacenEntry.AccountBranch,
						AccountNumber: &bacenEntry.AccountNumber,
						OwnerName:     &bacenEntry.OwnerName,
						OwnerTaxID:    &bacenEntry.OwnerTaxID,
					},
				})
			}
		}
	}

	// Find MISSING_BACEN entries
	for key, localEntry := range localMap {
		_, existsInBacen := bacenMap[key]

		if !existsInBacen {
			// Entry exists locally but not in Bacen - flag for manual review
			// DO NOT auto-delete - could be pending registration
			discrepancies = append(discrepancies, EntryDiscrepancy{
				Type:    DiscrepancyTypeMissingBacen,
				Key:     key,
				EntryID: localEntry.EntryID,
				Reason:  "Entry exists locally but not in Bacen DICT - needs manual review (may be pending registration)",
				LocalData: localEntry,
			})
		}
	}

	a.logger.WithFields(logrus.Fields{
		"total_discrepancies": len(discrepancies),
		"missing_local":       countDiscrepancyType(discrepancies, DiscrepancyTypeMissingLocal),
		"outdated_local":      countDiscrepancyType(discrepancies, DiscrepancyTypeOutdatedLocal),
		"missing_bacen":       countDiscrepancyType(discrepancies, DiscrepancyTypeMissingBacen),
	}).Info("Comparison completed")

	return discrepancies, nil
}

// isDataDifferent compares Bacen entry with local entry to detect differences
func isDataDifferent(bacen *BacenEntry, local *entities.Entry) bool {
	// Compare AccountBranch (local can be nil)
	if local.AccountBranch != nil && bacen.AccountBranch != *local.AccountBranch {
		return true
	}
	// Compare AccountNumber (local can be nil)
	if local.AccountNumber != nil && bacen.AccountNumber != *local.AccountNumber {
		return true
	}
	// Compare OwnerName (local can be nil)
	if local.OwnerName != nil && bacen.OwnerName != *local.OwnerName {
		return true
	}
	// Compare OwnerTaxID (local can be nil)
	if local.OwnerTaxID != nil && bacen.OwnerTaxID != *local.OwnerTaxID {
		return true
	}
	// Compare Status
	if bacen.Status != string(local.Status) {
		return true
	}
	return false
}

// countDiscrepancyType counts discrepancies by type
func countDiscrepancyType(discrepancies []EntryDiscrepancy, dtype string) int {
	count := 0
	for _, d := range discrepancies {
		if d.Type == dtype {
			count++
		}
	}
	return count
}

// GenerateSyncReportActivity generates an audit report for the sync operation
//
// This activity creates a detailed audit report documenting the sync process.
// The report is stored in the database for compliance and troubleshooting.
//
// Parameters:
// - result: VSyncResult containing sync statistics
//
// Returns:
// - string: Report ID for reference
//
// Implementation TODO:
// 1. Create SyncReport entity with:
//    - Sync timestamp
//    - Statistics (synced, created, updated, deleted, discrepancies)
//    - Status (COMPLETED, PARTIAL, FAILED)
//    - Error message (if any)
//    - Duration
// 2. Insert report into database (sync_reports table)
// 3. Optionally generate PDF/JSON report file
// 4. Return report ID for tracking
func (a *VSyncActivities) GenerateSyncReportActivity(
	ctx context.Context,
	syncID string,
	participantISPB string,
	syncType string,
	entriesFetched int,
	discrepancies []EntryDiscrepancy,
	created int,
	updated int,
	deleted int,
	duration time.Duration,
	syncTimestamp time.Time,
	status string,
	errorMessage string,
) (string, error) {
	a.logger.WithFields(logrus.Fields{
		"sync_id":         syncID,
		"participant_ispb": participantISPB,
		"sync_type":       syncType,
		"discrepancies":   len(discrepancies),
	}).Info("Generating sync audit report")

	// Create SyncReport entity
	var syncTypeEnum entities.SyncType
	if syncType == "FULL" {
		syncTypeEnum = entities.SyncTypeFull
	} else {
		syncTypeEnum = entities.SyncTypeIncremental
	}

	report := entities.NewSyncReport(syncID, syncTypeEnum, participantISPB)
	report.SyncTimestamp = syncTimestamp
	report.SetDuration(duration)

	// Set statistics
	report.EntriesFetched = entriesFetched
	report.EntriesCompared = entriesFetched
	report.DiscrepanciesFound = len(discrepancies)
	report.EntriesCreated = created
	report.EntriesUpdated = updated
	report.EntriesDeleted = deleted
	report.EntriesSynced = created + updated + deleted

	// Count discrepancies by type
	for _, d := range discrepancies {
		switch d.Type {
		case DiscrepancyTypeMissingLocal:
			report.DiscrepanciesMissingLocal++
		case DiscrepancyTypeOutdatedLocal:
			report.DiscrepanciesOutdatedLocal++
		case DiscrepancyTypeMissingBacen:
			report.DiscrepanciesMissingBacen++
		}
	}

	// Set status and error
	switch status {
	case "COMPLETED":
		report.Status = entities.SyncStatusCompleted
	case "PARTIAL":
		report.Status = entities.SyncStatusPartial
		if errorMessage != "" {
			report.SetPartial(errorMessage)
		}
	case "FAILED":
		report.Status = entities.SyncStatusFailed
		if errorMessage != "" {
			report.SetError(errorMessage, "VSYNC_ERROR")
		}
	}

	// Add metadata
	report.AddMetadata("total_discrepancies", len(discrepancies))
	report.AddMetadata("missing_local", report.DiscrepanciesMissingLocal)
	report.AddMetadata("outdated_local", report.DiscrepanciesOutdatedLocal)
	report.AddMetadata("missing_bacen", report.DiscrepanciesMissingBacen)

	// Insert report into database
	if err := a.reportRepo.Create(ctx, report); err != nil {
		a.logger.WithError(err).Error("Failed to create sync report")
		return "", fmt.Errorf("failed to create sync report: %w", err)
	}

	a.logger.WithFields(logrus.Fields{
		"report_id":       report.ID.String(),
		"sync_id":         syncID,
		"discrepancies":   len(discrepancies),
		"entries_synced":  report.EntriesSynced,
	}).Info("Sync report generated and stored")

	return report.ID.String(), nil
}

// Implementation Notes for Future Developers:
//
// 1. Bacen API Client (FetchBacenEntriesActivity):
//    - Use conn-bridge gRPC client to call RSFN Bridge
//    - Bridge will handle mTLS, XML signing, SOAP adapter
//    - Endpoint: bridge.QueryDICTEntries(ispb, syncType, lastSyncDate)
//    - Handle pagination: Bacen limits 1000 entries per request
//    - Retry on transient errors (network, rate limit)
//
// 2. Database Comparison (CompareEntriesActivity):
//    - Use EntryRepository.GetAll(ispb) to fetch local entries
//    - Create hash maps for O(n) comparison: map[key]Entry
//    - Compare field-by-field for OUTDATED_LOCAL detection
//    - Fields to compare: AccountBranch, AccountNumber, Status, OwnerName, etc.
//    - Use transaction for consistency
//
// 3. Report Generation (GenerateSyncReportActivity):
//    - Store report in PostgreSQL: sync_reports table
//    - Schema: id, sync_timestamp, entries_synced, entries_created, entries_updated,
//              entries_deleted, discrepancies, status, duration, error_message
//    - Optionally export to JSON/PDF for compliance team
//    - Retention: Keep reports for 7 years (Bacen requirement)
//
// 4. Error Handling:
//    - Bacen API errors: Retry with exponential backoff
//    - Database errors: Log and continue (partial sync is acceptable)
//    - Validation errors: Flag entry for manual review
//    - Never auto-delete entries missing in Bacen (could be pending registration)
//
// 5. Performance Optimization:
//    - Batch database operations (bulk insert/update)
//    - Use database indexes on Key, ISPB, UpdatedAt
//    - Consider parallel processing for large ISPB counts
//    - Heartbeat every 1000 entries for long-running activities
//
// 6. Compliance Considerations:
//    - All discrepancies must be logged for audit trail
//    - Generate detailed report for compliance team
//    - Alert on high discrepancy count (>1% threshold)
//    - Flag entries missing in Bacen for manual review (do NOT auto-delete)
//    - LGPD: Log only entry_id and key (not PII) in standard logs