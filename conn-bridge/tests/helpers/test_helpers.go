package helpers

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"testing"
	"time"

	pb "github.com/lbpay-lab/dict-contracts/gen/proto/bridge/v1"
	"github.com/lbpay-lab/conn-bridge/internal/domain/entities"
	grpcserver "github.com/lbpay-lab/conn-bridge/internal/grpc"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// SetupTestDB creates a test database using testcontainers
func SetupTestDB(t *testing.T) *sql.DB {
	ctx := context.Background()

	// Create PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "test_db",
			"POSTGRES_USER":     "test_user",
			"POSTGRES_PASSWORD": "test_password",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Get container host and port
	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Create connection string
	connStr := fmt.Sprintf(
		"host=%s port=%s user=test_user password=test_password dbname=test_db sslmode=disable",
		host, port.Port(),
	)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	// Verify connection
	err = db.Ping()
	require.NoError(t, err)

	// Cleanup
	t.Cleanup(func() {
		db.Close()
		container.Terminate(ctx)
	})

	return db
}

// SetupTestRedis creates a test Redis instance using testcontainers
func SetupTestRedis(t *testing.T) *redis.Client {
	ctx := context.Background()

	// Create Redis container
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Get container host and port
	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "6379")
	require.NoError(t, err)

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port.Port()),
	})

	// Verify connection
	_, err = client.Ping(ctx).Result()
	require.NoError(t, err)

	// Cleanup
	t.Cleanup(func() {
		client.Close()
		container.Terminate(ctx)
	})

	return client
}

// CleanupTest performs cleanup after tests
func CleanupTest(t *testing.T) {
	// This is called automatically via t.Cleanup() in setup functions
	t.Log("Test cleanup completed")
}

// LoadFixture loads a test fixture by name
func LoadFixture(t *testing.T, name string) interface{} {
	fixtures := map[string]interface{}{
		"valid_cpf_entry": CreateValidCPFEntry(),
		"valid_email_entry": CreateValidEmailEntry(),
		"valid_phone_entry": CreateValidPhoneEntry(),
	}

	fixture, ok := fixtures[name]
	require.True(t, ok, "Fixture %s not found", name)

	return fixture
}

// CreateValidCPFEntry creates a valid CPF entry for testing
func CreateValidCPFEntry() *entities.DictEntry {
	return &entities.DictEntry{
		Key:         "12345678901",
		Type:        entities.KeyTypeCPF,
		Participant: "60701190",
		Account: entities.Account{
			ISPB:        "60701190",
			Branch:      "0001",
			Number:      "123456",
			Type:        entities.AccountTypeChecking,
			OpeningDate: time.Now(),
		},
		Owner: entities.Owner{
			Type:     entities.OwnerTypePerson,
			Document: "12345678901",
			Name:     "Test User CPF",
		},
		Status:    entities.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateValidEmailEntry creates a valid email entry for testing
func CreateValidEmailEntry() *entities.DictEntry {
	return &entities.DictEntry{
		Key:         "test@example.com",
		Type:        entities.KeyTypeEmail,
		Participant: "60701190",
		Account: entities.Account{
			ISPB:        "60701190",
			Branch:      "0001",
			Number:      "654321",
			Type:        entities.AccountTypeSavings,
			OpeningDate: time.Now(),
		},
		Owner: entities.Owner{
			Type:     entities.OwnerTypePerson,
			Document: "98765432100",
			Name:     "Test User Email",
		},
		Status:    entities.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateValidPhoneEntry creates a valid phone entry for testing
func CreateValidPhoneEntry() *entities.DictEntry {
	return &entities.DictEntry{
		Key:         "+5511987654321",
		Type:        entities.KeyTypePhone,
		Participant: "60701190",
		Account: entities.Account{
			ISPB:        "60701190",
			Branch:      "0001",
			Number:      "111222",
			Type:        entities.AccountTypePayment,
			OpeningDate: time.Now(),
		},
		Owner: entities.Owner{
			Type:     entities.OwnerTypePerson,
			Document: "11122233344",
			Name:     "Test User Phone",
		},
		Status:    entities.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// WaitForCondition waits for a condition to be true or timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, interval time.Duration) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(interval)
	}

	t.Fatal("Timeout waiting for condition")
}

// AssertEventuallyTrue asserts that a condition becomes true within a timeout
func AssertEventuallyTrue(t *testing.T, condition func() bool, timeout time.Duration, msgAndArgs ...interface{}) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	require.Fail(t, "Condition did not become true within timeout", msgAndArgs...)
}

// CreateTestContext creates a context with a reasonable timeout for tests
func CreateTestContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// CreateTestContextWithTimeout creates a context with a custom timeout
func CreateTestContextWithTimeout(t *testing.T, timeout time.Duration) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)
	return ctx
}

// RandomString generates a random string for testing
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// GenerateTestCPF generates a test CPF number
func GenerateTestCPF() string {
	return fmt.Sprintf("%011d", time.Now().Unix()%100000000000)
}

// GenerateTestEmail generates a test email
func GenerateTestEmail() string {
	return fmt.Sprintf("test_%d@example.com", time.Now().UnixNano())
}

// GenerateTestPhone generates a test phone number
func GenerateTestPhone() string {
	return fmt.Sprintf("+5511%09d", time.Now().Unix()%1000000000)
}

// MockSOAPClient is a mock implementation of SOAPClient for testing
type MockSOAPClient struct{}

func (m *MockSOAPClient) SendSOAPRequest(ctx context.Context, endpoint string, soapEnvelope []byte) ([]byte, error) {
	return []byte(`<soap:Envelope><soap:Body><Response>OK</Response></soap:Body></soap:Envelope>`), nil
}

func (m *MockSOAPClient) BuildSOAPEnvelope(bodyXML string, signedXML string) ([]byte, error) {
	return []byte(fmt.Sprintf(`<soap:Envelope><soap:Body>%s</soap:Body></soap:Envelope>`, bodyXML)), nil
}

func (m *MockSOAPClient) ParseSOAPResponse(soapResponse []byte) ([]byte, error) {
	return soapResponse, nil
}

func (m *MockSOAPClient) HealthCheck(ctx context.Context) error {
	return nil
}

// MockXMLSigner is a mock implementation of XMLSigner for testing
type MockXMLSigner struct{}

func (m *MockXMLSigner) SignXML(ctx context.Context, xmlData string) (string, error) {
	return xmlData + "<Signature>MOCK_SIGNATURE</Signature>", nil
}

func (m *MockXMLSigner) HealthCheck(ctx context.Context) error {
	return nil
}

// SetupTestClient creates a gRPC test client connected to a test server
func SetupTestClient(t *testing.T) (pb.BridgeServiceClient, func()) {
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Find available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Create mock dependencies
	mockSOAPClient := &MockSOAPClient{}
	mockXMLSigner := &MockXMLSigner{}

	// Create and start gRPC server
	server := grpcserver.NewServer(logger, port, mockSOAPClient, mockXMLSigner)

	// Start server in background
	go func() {
		if err := server.Start(); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to be ready
	time.Sleep(500 * time.Millisecond)

	// Create client connection
	conn, err := grpc.Dial(
		fmt.Sprintf("localhost:%d", port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	client := pb.NewBridgeServiceClient(conn)

	// Cleanup function
	cleanup := func() {
		conn.Close()
		server.Stop()
	}

	return client, cleanup
}
