package testhelpers

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// E2EEnvironment holds dependencies for E2E tests
type E2EEnvironment struct {
	Ctx        context.Context
	T          *testing.T
	CoreURL    string
	ConnectURL string
	BridgeURL  string
	HTTPClient *http.Client
}

// SetupE2ETest initializes E2E test environment
// Assumes services are already running via docker-compose
func SetupE2ETest(t *testing.T) *E2EEnvironment {
	ctx := context.Background()

	env := &E2EEnvironment{
		Ctx:        ctx,
		T:          t,
		CoreURL:    getEnvOrDefault("CORE_DICT_URL", "http://localhost:8080"),
		ConnectURL: getEnvOrDefault("CONN_DICT_URL", "http://localhost:8081"),
		BridgeURL:  getEnvOrDefault("CONN_BRIDGE_URL", "http://localhost:8082"),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Wait for services to be healthy
	t.Log("Waiting for services to be healthy...")
	WaitForHealthy(t, env.CoreURL+"/health", 60*time.Second)
	WaitForHealthy(t, env.ConnectURL+"/health", 60*time.Second)
	WaitForHealthy(t, env.BridgeURL+"/health", 60*time.Second)

	return env
}

// WaitForHealthy waits for a service to be healthy
func WaitForHealthy(t *testing.T, url string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 5 * time.Second}

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			t.Logf("Service %s is healthy", url)
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}

	require.Fail(t, fmt.Sprintf("Service %s did not become healthy within %v", url, timeout))
}

func getEnvOrDefault(key, defaultValue string) string {
	// In real implementation, would read from os.Getenv
	return defaultValue
}
