package grpc

import (
	"context"
	"fmt"
	"time"

	bridgev1 "github.com/lbpay-lab/dict-contracts/gen/proto/bridge/v1"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// BridgeClient wraps the gRPC client for Bridge service
type BridgeClient struct {
	client bridgev1.BridgeServiceClient
	conn   *grpc.ClientConn
	logger *logrus.Logger
	tracer trace.Tracer
}

// BridgeClientConfig holds configuration for Bridge client
type BridgeClientConfig struct {
	Address        string
	ConnectTimeout time.Duration
	RequestTimeout time.Duration
}

// NewBridgeClient creates a new Bridge gRPC client
func NewBridgeClient(config *BridgeClientConfig, logger *logrus.Logger) (*BridgeClient, error) {
	if config.ConnectTimeout == 0 {
		config.ConnectTimeout = 10 * time.Second
	}
	if config.RequestTimeout == 0 {
		config.RequestTimeout = 30 * time.Second
	}

	logger.WithField("address", config.Address).Info("Connecting to Bridge service")

	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()

	// TODO: Add TLS credentials in production
	conn, err := grpc.DialContext(ctx, config.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Bridge service at %s: %w", config.Address, err)
	}

	client := bridgev1.NewBridgeServiceClient(conn)
	tracer := otel.Tracer("conn-dict/bridge-client")

	logger.Info("Successfully connected to Bridge service")

	return &BridgeClient{
		client: client,
		conn:   conn,
		logger: logger,
		tracer: tracer,
	}, nil
}

// Close closes the Bridge client connection
func (c *BridgeClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// CreateEntry calls Bridge to create a new DICT entry
func (c *BridgeClient) CreateEntry(ctx context.Context, req *bridgev1.CreateEntryRequest) (*bridgev1.CreateEntryResponse, error) {
	ctx, span := c.tracer.Start(ctx, "BridgeClient.CreateEntry")
	defer span.End()

	c.logger.WithFields(logrus.Fields{
		"key_type":       req.Key.KeyType.String(),
		"idempotency_key": req.IdempotencyKey,
		"request_id":     req.RequestId,
	}).Debug("Calling Bridge CreateEntry")

	resp, err := c.client.CreateEntry(ctx, req)
	if err != nil {
		c.logger.WithError(err).Error("Bridge CreateEntry failed")
		return nil, fmt.Errorf("bridge CreateEntry failed: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"entry_id":    resp.EntryId,
		"external_id": resp.ExternalId,
		"status":      resp.Status.String(),
	}).Info("Bridge CreateEntry succeeded")

	return resp, nil
}

// GetEntry calls Bridge to retrieve a DICT entry
func (c *BridgeClient) GetEntry(ctx context.Context, req *bridgev1.GetEntryRequest) (*bridgev1.GetEntryResponse, error) {
	ctx, span := c.tracer.Start(ctx, "BridgeClient.GetEntry")
	defer span.End()

	c.logger.WithField("request_id", req.RequestId).Debug("Calling Bridge GetEntry")

	resp, err := c.client.GetEntry(ctx, req)
	if err != nil {
		c.logger.WithError(err).Error("Bridge GetEntry failed")
		return nil, fmt.Errorf("bridge GetEntry failed: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"entry_id": resp.EntryId,
		"found":    resp.Found,
	}).Debug("Bridge GetEntry succeeded")

	return resp, nil
}

// UpdateEntry calls Bridge to update a DICT entry
func (c *BridgeClient) UpdateEntry(ctx context.Context, req *bridgev1.UpdateEntryRequest) (*bridgev1.UpdateEntryResponse, error) {
	ctx, span := c.tracer.Start(ctx, "BridgeClient.UpdateEntry")
	defer span.End()

	c.logger.WithFields(logrus.Fields{
		"entry_id":        req.EntryId,
		"idempotency_key": req.IdempotencyKey,
		"request_id":      req.RequestId,
	}).Debug("Calling Bridge UpdateEntry")

	resp, err := c.client.UpdateEntry(ctx, req)
	if err != nil {
		c.logger.WithError(err).Error("Bridge UpdateEntry failed")
		return nil, fmt.Errorf("bridge UpdateEntry failed: %w", err)
	}

	c.logger.WithField("entry_id", resp.EntryId).Info("Bridge UpdateEntry succeeded")

	return resp, nil
}

// DeleteEntry calls Bridge to delete a DICT entry
func (c *BridgeClient) DeleteEntry(ctx context.Context, req *bridgev1.DeleteEntryRequest) (*bridgev1.DeleteEntryResponse, error) {
	ctx, span := c.tracer.Start(ctx, "BridgeClient.DeleteEntry")
	defer span.End()

	c.logger.WithFields(logrus.Fields{
		"entry_id":        req.EntryId,
		"idempotency_key": req.IdempotencyKey,
		"request_id":      req.RequestId,
	}).Debug("Calling Bridge DeleteEntry")

	resp, err := c.client.DeleteEntry(ctx, req)
	if err != nil {
		c.logger.WithError(err).Error("Bridge DeleteEntry failed")
		return nil, fmt.Errorf("bridge DeleteEntry failed: %w", err)
	}

	c.logger.WithField("deleted", resp.Deleted).Info("Bridge DeleteEntry succeeded")

	return resp, nil
}

// SearchEntries calls Bridge to search for DICT entries with filters and pagination
func (c *BridgeClient) SearchEntries(ctx context.Context, req *bridgev1.SearchEntriesRequest) (*bridgev1.SearchEntriesResponse, error) {
	ctx, span := c.tracer.Start(ctx, "BridgeClient.SearchEntries")
	defer span.End()

	c.logger.WithFields(logrus.Fields{
		"ispb":       req.Ispb,
		"page_size":  req.PageSize,
		"page_token": req.PageToken,
		"request_id": req.RequestId,
	}).Debug("Calling Bridge SearchEntries")

	resp, err := c.client.SearchEntries(ctx, req)
	if err != nil {
		c.logger.WithError(err).Error("Bridge SearchEntries failed")
		return nil, fmt.Errorf("bridge SearchEntries failed: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"entries_count":    len(resp.Entries),
		"total_count":      resp.TotalCount,
		"next_page_token":  resp.NextPageToken,
	}).Info("Bridge SearchEntries succeeded")

	return resp, nil
}
