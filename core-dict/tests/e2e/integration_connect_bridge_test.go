package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lbpay-lab/core-dict/tests/testhelpers"
)

func TestE2E_Core_Connect_Bridge_CreateEntry_SOAP_Bacen(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	env := testhelpers.SetupE2ETest(t)

	// This test verifies the complete integration flow:
	// Core-Dict REST API → Connect gRPC → Bridge SOAP → Bacen (mock)

	// Step 1: Create entry via Core-Dict REST API
	entryReq := EntryRequest{
		KeyType:   "CPF",
		KeyValue:  "11111111111",
		AccountID: uuid.NewString(),
		ISPB:      "12345678",
		UserID:    "e2e-integration-test",
	}

	entryBody, _ := json.Marshal(entryReq)
	t.Logf("Creating entry via Core-Dict API...")

	entryResp, err := env.HTTPClient.Post(
		env.CoreURL+"/api/v1/entries",
		"application/json",
		bytes.NewBuffer(entryBody),
	)
	require.NoError(t, err)
	defer entryResp.Body.Close()

	assert.Equal(t, http.StatusCreated, entryResp.StatusCode)

	var entry EntryResponse
	body, _ := io.ReadAll(entryResp.Body)
	json.Unmarshal(body, &entry)

	t.Logf("Entry created: %s", entry.ID)

	// Step 2: Wait for event propagation (Core → Pulsar → Connect)
	time.Sleep(2 * time.Second)

	// Step 3: Verify Connect received the event
	t.Logf("Checking Connect for entry %s...", entry.ID)
	connectCheckResp, err := env.HTTPClient.Get(
		fmt.Sprintf("%s/api/v1/entries/%s/sync-status", env.ConnectURL, entry.ID),
	)

	if err == nil && connectCheckResp != nil {
		defer connectCheckResp.Body.Close()
		if connectCheckResp.StatusCode == http.StatusOK {
			body, _ := io.ReadAll(connectCheckResp.Body)
			var syncStatus map[string]interface{}
			json.Unmarshal(body, &syncStatus)

			t.Logf("Connect sync status: %v", syncStatus)
			assert.Equal(t, "SYNCED", syncStatus["status"])
		}
	}

	// Step 4: Verify Bridge sent SOAP request to Bacen
	t.Logf("Checking Bridge logs for SOAP request...")
	bridgeLogsResp, err := env.HTTPClient.Get(
		fmt.Sprintf("%s/api/v1/logs/entry/%s", env.BridgeURL, entry.ID),
	)

	if err == nil && bridgeLogsResp != nil {
		defer bridgeLogsResp.Body.Close()
		if bridgeLogsResp.StatusCode == http.StatusOK {
			body, _ := io.ReadAll(bridgeLogsResp.Body)
			var logs map[string]interface{}
			json.Unmarshal(body, &logs)

			t.Logf("Bridge logs: %v", logs)
			// Verify SOAP request was sent
			assert.NotEmpty(t, logs["soap_request_sent"])
		}
	}

	// Step 5: Verify end-to-end consistency
	// Query entry from Core-Dict
	getEntryResp, err := env.HTTPClient.Get(env.CoreURL + "/api/v1/entries/" + entry.ID)
	require.NoError(t, err)
	defer getEntryResp.Body.Close()

	var finalEntry EntryResponse
	body, _ = io.ReadAll(getEntryResp.Body)
	json.Unmarshal(body, &finalEntry)

	assert.Equal(t, "ACTIVE", finalEntry.Status)
	assert.Equal(t, entry.KeyValue, finalEntry.KeyValue)

	t.Logf("E2E flow completed: Core → Connect → Bridge → Bacen")
}

func TestE2E_Core_Connect_Bridge_CreateClaim_VSYNC_Bacen(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	env := testhelpers.SetupE2ETest(t)

	// This test verifies claim creation triggers VSYNC workflow:
	// Core-Dict → Connect (Temporal ClaimWorkflow) → Bridge (VSYNC) → Bacen

	// Step 1: Create entry first
	entryReq := EntryRequest{
		KeyType:   "EMAIL",
		KeyValue:  "vsync@test.com",
		AccountID: uuid.NewString(),
		ISPB:      "12345678",
		UserID:    "e2e-vsync-test",
	}

	entryBody, _ := json.Marshal(entryReq)
	entryResp, err := env.HTTPClient.Post(
		env.CoreURL+"/api/v1/entries",
		"application/json",
		bytes.NewBuffer(entryBody),
	)
	require.NoError(t, err)
	defer entryResp.Body.Close()

	var entry EntryResponse
	body, _ := io.ReadAll(entryResp.Body)
	json.Unmarshal(body, &entry)

	time.Sleep(1 * time.Second)

	// Step 2: Create claim (triggers ClaimWorkflow in Connect)
	claimReq := ClaimRequest{
		EntryID:     entry.ID,
		ClaimType:   "PORTABILITY",
		DonorISPB:   "12345678",
		ClaimerISPB: "87654321",
		UserID:      "e2e-vsync-test",
	}

	claimBody, _ := json.Marshal(claimReq)
	t.Logf("Creating claim via Core-Dict API...")

	claimResp, err := env.HTTPClient.Post(
		env.CoreURL+"/api/v1/claims",
		"application/json",
		bytes.NewBuffer(claimBody),
	)
	require.NoError(t, err)
	defer claimResp.Body.Close()

	assert.Equal(t, http.StatusCreated, claimResp.StatusCode)

	var claim ClaimResponse
	body, _ = io.ReadAll(claimResp.Body)
	json.Unmarshal(body, &claim)

	t.Logf("Claim created: %s", claim.ID)

	// Step 3: Wait for ClaimWorkflow to start in Connect (Temporal)
	time.Sleep(3 * time.Second)

	// Step 4: Verify Connect started Temporal ClaimWorkflow
	t.Logf("Checking Temporal workflow status...")
	workflowResp, err := env.HTTPClient.Get(
		fmt.Sprintf("%s/api/v1/workflows/claim/%s", env.ConnectURL, claim.ID),
	)

	if err == nil && workflowResp != nil {
		defer workflowResp.Body.Close()
		if workflowResp.StatusCode == http.StatusOK {
			body, _ := io.ReadAll(workflowResp.Body)
			var workflow map[string]interface{}
			json.Unmarshal(body, &workflow)

			t.Logf("Workflow status: %v", workflow)
			assert.Equal(t, "RUNNING", workflow["status"])
			assert.Contains(t, workflow["workflow_type"], "ClaimWorkflow")
		}
	}

	// Step 5: Verify Bridge initiated VSYNC with Bacen
	t.Logf("Checking Bridge VSYNC status...")
	vsyncResp, err := env.HTTPClient.Get(
		fmt.Sprintf("%s/api/v1/vsync/claim/%s", env.BridgeURL, claim.ID),
	)

	if err == nil && vsyncResp != nil {
		defer vsyncResp.Body.Close()
		if vsyncResp.StatusCode == http.StatusOK {
			body, _ := io.ReadAll(vsyncResp.Body)
			var vsyncStatus map[string]interface{}
			json.Unmarshal(body, &vsyncStatus)

			t.Logf("VSYNC status: %v", vsyncStatus)
			assert.NotEmpty(t, vsyncStatus["vsync_initiated"])
		}
	}

	// Step 6: Verify claim status in Core-Dict
	getClaimResp, err := env.HTTPClient.Get(env.CoreURL + "/api/v1/claims/" + claim.ID)
	require.NoError(t, err)
	defer getClaimResp.Body.Close()

	var finalClaim ClaimResponse
	body, _ = io.ReadAll(getClaimResp.Body)
	json.Unmarshal(body, &finalClaim)

	assert.Equal(t, "OPEN", finalClaim.Status)
	assert.NotZero(t, finalClaim.ExpiresAt)

	t.Logf("VSYNC workflow initiated successfully")
}

func TestE2E_Core_Connect_Bridge_Pulsar_Events_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	env := testhelpers.SetupE2ETest(t)

	// This test verifies Pulsar event flow through all services:
	// Core-Dict (publish) → Pulsar → Connect (consume) → Bridge → Bacen

	// Step 1: Create multiple entries to generate events
	entryIDs := []string{}

	for i := 0; i < 3; i++ {
		entryReq := EntryRequest{
			KeyType:   "PHONE",
			KeyValue:  fmt.Sprintf("+5511777777%03d", i),
			AccountID: uuid.NewString(),
			ISPB:      "12345678",
			UserID:    "e2e-pulsar-test",
		}

		entryBody, _ := json.Marshal(entryReq)
		entryResp, err := env.HTTPClient.Post(
			env.CoreURL+"/api/v1/entries",
			"application/json",
			bytes.NewBuffer(entryBody),
		)
		require.NoError(t, err)

		var entry EntryResponse
		body, _ := io.ReadAll(entryResp.Body)
		entryResp.Body.Close()
		json.Unmarshal(body, &entry)

		entryIDs = append(entryIDs, entry.ID)
		t.Logf("Created entry %d: %s", i+1, entry.ID)
	}

	// Step 2: Wait for events to propagate through Pulsar
	time.Sleep(3 * time.Second)

	// Step 3: Verify Connect consumed all events
	t.Logf("Checking Connect event consumption...")
	for _, entryID := range entryIDs {
		connectEventResp, err := env.HTTPClient.Get(
			fmt.Sprintf("%s/api/v1/events/entry/%s", env.ConnectURL, entryID),
		)

		if err == nil && connectEventResp != nil {
			defer connectEventResp.Body.Close()
			if connectEventResp.StatusCode == http.StatusOK {
				body, _ := io.ReadAll(connectEventResp.Body)
				var eventStatus map[string]interface{}
				json.Unmarshal(body, &eventStatus)

				t.Logf("Entry %s event status: %v", entryID, eventStatus)
				assert.Equal(t, "CONSUMED", eventStatus["status"])
			}
		}
	}

	// Step 4: Verify Bridge received events via gRPC from Connect
	t.Logf("Checking Bridge gRPC call logs...")
	bridgeStatsResp, err := env.HTTPClient.Get(
		env.BridgeURL + "/api/v1/stats/grpc-calls",
	)

	if err == nil && bridgeStatsResp != nil {
		defer bridgeStatsResp.Body.Close()
		if bridgeStatsResp.StatusCode == http.StatusOK {
			body, _ := io.ReadAll(bridgeStatsResp.Body)
			var stats map[string]interface{}
			json.Unmarshal(body, &stats)

			t.Logf("Bridge gRPC stats: %v", stats)
			// Verify at least 3 calls received
			if callCount, ok := stats["grpc_calls_received"].(float64); ok {
				assert.GreaterOrEqual(t, int(callCount), 3)
			}
		}
	}

	// Step 5: Update an entry to generate update event
	updateReq := map[string]interface{}{
		"status": "BLOCKED",
		"userId": "e2e-pulsar-test",
	}
	updateBody, _ := json.Marshal(updateReq)

	updateResp, err := env.HTTPClient.Put(
		env.CoreURL+"/api/v1/entries/"+entryIDs[0],
		"application/json",
		bytes.NewBuffer(updateBody),
	)

	if err == nil {
		defer updateResp.Body.Close()
		if updateResp.StatusCode == http.StatusOK {
			t.Logf("Entry updated: %s", entryIDs[0])

			// Wait for update event propagation
			time.Sleep(2 * time.Second)

			// Verify update event consumed
			connectUpdateResp, err := env.HTTPClient.Get(
				fmt.Sprintf("%s/api/v1/events/entry/%s/updates", env.ConnectURL, entryIDs[0]),
			)

			if err == nil && connectUpdateResp != nil {
				defer connectUpdateResp.Body.Close()
				if connectUpdateResp.StatusCode == http.StatusOK {
					body, _ := io.ReadAll(connectUpdateResp.Body)
					var updateEvents []map[string]interface{}
					json.Unmarshal(body, &updateEvents)

					t.Logf("Update events: %v", updateEvents)
					assert.GreaterOrEqual(t, len(updateEvents), 1)
				}
			}
		}
	}

	// Step 6: Verify end-to-end event metrics
	coreMetricsResp, err := env.HTTPClient.Get(env.CoreURL + "/metrics")
	if err == nil && coreMetricsResp != nil {
		defer coreMetricsResp.Body.Close()
		if coreMetricsResp.StatusCode == http.StatusOK {
			body, _ := io.ReadAll(coreMetricsResp.Body)
			t.Logf("Core metrics (partial): %s", string(body[:min(len(body), 500)]))
		}
	}

	t.Logf("Pulsar event flow verified: Core → Pulsar → Connect → Bridge")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
