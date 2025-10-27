package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lbpay-lab/core-dict/tests/testhelpers"
)

func TestE2E_Performance_CreateEntry_1000TPS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	env := testhelpers.SetupE2ETest(t)

	// Performance target: 1000 TPS (Transactions Per Second)
	// Test duration: 10 seconds
	// Expected: 10,000 requests completed

	duration := 10 * time.Second
	targetTPS := 1000
	expectedRequests := targetTPS * int(duration.Seconds())

	// Counters
	var (
		successCount int64
		errorCount   int64
		totalLatency int64
	)

	// Start time
	startTime := time.Now()
	endTime := startTime.Add(duration)

	// Worker pool
	workerCount := 50
	var wg sync.WaitGroup

	t.Logf("Starting performance test: Target %d TPS for %v", targetTPS, duration)
	t.Logf("Workers: %d", workerCount)

	// Rate limiter channel
	rateLimiter := make(chan struct{}, targetTPS/10) // Send in batches
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			if time.Now().After(endTime) {
				close(rateLimiter)
				return
			}
			for i := 0; i < targetTPS/10; i++ {
				select {
				case rateLimiter <- struct{}{}:
				default:
					return
				}
			}
		}
	}()

	// Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			client := &http.Client{
				Timeout: 5 * time.Second,
				Transport: &http.Transport{
					MaxIdleConnsPerHost: 100,
				},
			}

			for range rateLimiter {
				reqStartTime := time.Now()

				// Create unique entry
				entryReq := EntryRequest{
					KeyType:   "PHONE",
					KeyValue:  fmt.Sprintf("+5511%d%09d", workerID, time.Now().UnixNano()%1000000000),
					AccountID: uuid.NewString(),
					ISPB:      "12345678",
					UserID:    fmt.Sprintf("perf-test-worker-%d", workerID),
				}

				reqBody, _ := json.Marshal(entryReq)

				resp, err := client.Post(
					env.CoreURL+"/api/v1/entries",
					"application/json",
					bytes.NewBuffer(reqBody),
				)

				latency := time.Since(reqStartTime).Milliseconds()
				atomic.AddInt64(&totalLatency, latency)

				if err != nil {
					atomic.AddInt64(&errorCount, 1)
					continue
				}

				if resp.StatusCode == http.StatusCreated {
					atomic.AddInt64(&successCount, 1)
					io.Copy(io.Discard, resp.Body)
				} else {
					atomic.AddInt64(&errorCount, 1)
				}
				resp.Body.Close()
			}
		}(i)
	}

	// Wait for all workers to complete
	wg.Wait()

	// Calculate metrics
	actualDuration := time.Since(startTime)
	actualTPS := float64(successCount) / actualDuration.Seconds()
	avgLatency := float64(totalLatency) / float64(successCount+errorCount)
	errorRate := float64(errorCount) / float64(successCount+errorCount) * 100

	// Report results
	t.Logf("=== Performance Test Results ===")
	t.Logf("Duration: %v", actualDuration)
	t.Logf("Total Requests: %d", successCount+errorCount)
	t.Logf("Successful: %d", successCount)
	t.Logf("Errors: %d", errorCount)
	t.Logf("Actual TPS: %.2f", actualTPS)
	t.Logf("Avg Latency: %.2f ms", avgLatency)
	t.Logf("Error Rate: %.2f%%", errorRate)

	// Assertions
	assert.Greater(t, successCount, int64(expectedRequests*0.9), "Should achieve at least 90% of target requests")
	assert.Greater(t, actualTPS, float64(targetTPS)*0.9, "Should achieve at least 90% of target TPS")
	assert.Less(t, errorRate, 5.0, "Error rate should be less than 5%")
	assert.Less(t, avgLatency, 100.0, "Average latency should be less than 100ms")

	// P95 and P99 latency would require histogram tracking
	// For simplicity, we just check average latency
}

func TestE2E_Performance_Concurrent_Claims_100Parallel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	env := testhelpers.SetupE2ETest(t)

	// Performance test: Create 100 claims in parallel
	// Verify system can handle concurrent claim creation

	parallelClaims := 100

	// Step 1: Create entries first (one per claim)
	t.Logf("Creating %d entries for claims...", parallelClaims)

	type EntryClaimPair struct {
		EntryID string
		Index   int
	}

	entries := make([]EntryClaimPair, parallelClaims)
	var entryWg sync.WaitGroup
	var entryCreationErrors int64

	for i := 0; i < parallelClaims; i++ {
		entryWg.Add(1)
		go func(idx int) {
			defer entryWg.Done()

			entryReq := EntryRequest{
				KeyType:   "EMAIL",
				KeyValue:  fmt.Sprintf("claim-perf-%d@test.com", idx),
				AccountID: uuid.NewString(),
				ISPB:      "12345678",
				UserID:    fmt.Sprintf("perf-claim-test-%d", idx),
			}

			reqBody, _ := json.Marshal(entryReq)

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Post(
				env.CoreURL+"/api/v1/entries",
				"application/json",
				bytes.NewBuffer(reqBody),
			)

			if err != nil {
				atomic.AddInt64(&entryCreationErrors, 1)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusCreated {
				var entry EntryResponse
				body, _ := io.ReadAll(resp.Body)
				json.Unmarshal(body, &entry)
				entries[idx] = EntryClaimPair{EntryID: entry.ID, Index: idx}
			} else {
				atomic.AddInt64(&entryCreationErrors, 1)
			}
		}(i)
	}

	entryWg.Wait()

	require.Equal(t, int64(0), entryCreationErrors, "All entries should be created successfully")

	// Wait for entries to propagate
	time.Sleep(2 * time.Second)

	// Step 2: Create claims in parallel
	t.Logf("Creating %d claims in parallel...", parallelClaims)

	var (
		claimSuccessCount int64
		claimErrorCount   int64
		totalClaimLatency int64
	)

	startTime := time.Now()
	var claimWg sync.WaitGroup

	for i := 0; i < parallelClaims; i++ {
		if entries[i].EntryID == "" {
			continue // Skip if entry creation failed
		}

		claimWg.Add(1)
		go func(idx int, entryID string) {
			defer claimWg.Done()

			reqStartTime := time.Now()

			claimReq := ClaimRequest{
				EntryID:     entryID,
				ClaimType:   "OWNERSHIP",
				DonorISPB:   "12345678",
				ClaimerISPB: "87654321",
				UserID:      fmt.Sprintf("perf-claim-test-%d", idx),
			}

			reqBody, _ := json.Marshal(claimReq)

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Post(
				env.CoreURL+"/api/v1/claims",
				"application/json",
				bytes.NewBuffer(reqBody),
			)

			latency := time.Since(reqStartTime).Milliseconds()
			atomic.AddInt64(&totalClaimLatency, latency)

			if err != nil {
				atomic.AddInt64(&claimErrorCount, 1)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusCreated {
				atomic.AddInt64(&claimSuccessCount, 1)
			} else {
				atomic.AddInt64(&claimErrorCount, 1)
				body, _ := io.ReadAll(resp.Body)
				t.Logf("Claim creation failed for entry %s: %s", entryID, string(body))
			}
		}(i, entries[i].EntryID)
	}

	claimWg.Wait()

	// Calculate metrics
	totalDuration := time.Since(startTime)
	avgClaimLatency := float64(totalClaimLatency) / float64(claimSuccessCount+claimErrorCount)
	claimErrorRate := float64(claimErrorCount) / float64(claimSuccessCount+claimErrorCount) * 100

	// Report results
	t.Logf("=== Concurrent Claims Test Results ===")
	t.Logf("Total Duration: %v", totalDuration)
	t.Logf("Successful Claims: %d", claimSuccessCount)
	t.Logf("Failed Claims: %d", claimErrorCount)
	t.Logf("Avg Claim Creation Latency: %.2f ms", avgClaimLatency)
	t.Logf("Claim Error Rate: %.2f%%", claimErrorRate)

	// Assertions
	assert.Greater(t, claimSuccessCount, int64(parallelClaims*0.95), "Should create at least 95% of claims successfully")
	assert.Less(t, claimErrorRate, 5.0, "Error rate should be less than 5%")
	assert.Less(t, avgClaimLatency, 500.0, "Average latency should be less than 500ms")
	assert.Less(t, totalDuration.Seconds(), 30.0, "Should complete within 30 seconds")

	// Step 3: Verify all claims are in correct state
	time.Sleep(2 * time.Second)

	t.Logf("Verifying claim states...")
	var verifyWg sync.WaitGroup
	var validClaims int64

	client := &http.Client{Timeout: 5 * time.Second}

	for i := 0; i < int(claimSuccessCount); i++ {
		if entries[i].EntryID == "" {
			continue
		}

		verifyWg.Add(1)
		go func(entryID string) {
			defer verifyWg.Done()

			// Query claims for this entry
			resp, err := client.Get(
				fmt.Sprintf("%s/api/v1/entries/%s/claims", env.CoreURL, entryID),
			)

			if err == nil && resp.StatusCode == http.StatusOK {
				defer resp.Body.Close()
				var claims []ClaimResponse
				body, _ := io.ReadAll(resp.Body)
				json.Unmarshal(body, &claims)

				if len(claims) > 0 && claims[0].Status == "OPEN" {
					atomic.AddInt64(&validClaims, 1)
				}
			}
		}(entries[i].EntryID)
	}

	verifyWg.Wait()

	t.Logf("Valid claims in OPEN state: %d", validClaims)
	assert.Greater(t, validClaims, int64(claimSuccessCount*0.9), "At least 90% of claims should be in valid state")

	// Final summary
	t.Logf("=== Performance Test Summary ===")
	t.Logf("✓ Created %d entries", parallelClaims)
	t.Logf("✓ Created %d claims concurrently", claimSuccessCount)
	t.Logf("✓ Average latency: %.2f ms", avgClaimLatency)
	t.Logf("✓ Total duration: %v", totalDuration)
	t.Logf("✓ System handled %d concurrent operations successfully", parallelClaims)
}
