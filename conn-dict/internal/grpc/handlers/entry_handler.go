package handlers

import (
	"context"

	bridgev1 "github.com/lbpay-lab/dict-contracts/gen/proto/bridge/v1"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// EntryUseCase interface for business logic
type EntryUseCase interface {
	CreateEntry(ctx context.Context, req *bridgev1.CreateEntryRequest) (*bridgev1.CreateEntryResponse, error)
	GetEntry(ctx context.Context, req *bridgev1.GetEntryRequest) (*bridgev1.GetEntryResponse, error)
	UpdateEntry(ctx context.Context, req *bridgev1.UpdateEntryRequest) (*bridgev1.UpdateEntryResponse, error)
	DeleteEntry(ctx context.Context, req *bridgev1.DeleteEntryRequest) (*bridgev1.DeleteEntryResponse, error)
}

// EntryHandler implements the Bridge service gRPC handlers for Connect
type EntryHandler struct {
	bridgev1.UnimplementedBridgeServiceServer
	useCase EntryUseCase
	logger  *logrus.Logger
	tracer  trace.Tracer
}

// NewEntryHandler creates a new EntryHandler
func NewEntryHandler(useCase EntryUseCase, logger *logrus.Logger, tracer trace.Tracer) *EntryHandler {
	return &EntryHandler{
		useCase: useCase,
		logger:  logger,
		tracer:  tracer,
	}
}

// CreateEntry handles the CreateEntry RPC call
func (h *EntryHandler) CreateEntry(ctx context.Context, req *bridgev1.CreateEntryRequest) (*bridgev1.CreateEntryResponse, error) {
	ctx, span := h.tracer.Start(ctx, "EntryHandler.CreateEntry")
	defer span.End()

	h.logger.WithFields(logrus.Fields{
		"key_type":        req.Key.GetKeyType().String(),
		"key_value":       req.Key.GetKeyValue(),
		"ispb":            req.Account.GetIspb(),
		"request_id":      req.RequestId,
		"idempotency_key": req.IdempotencyKey,
	}).Info("CreateEntry RPC called")

	// Validate request
	if err := h.validateCreateEntryRequest(req); err != nil {
		h.logger.WithError(err).Warn("Invalid CreateEntry request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Call use case
	resp, err := h.useCase.CreateEntry(ctx, req)
	if err != nil {
		h.logger.WithError(err).Error("CreateEntry use case failed")
		return nil, h.mapError(err)
	}

	h.logger.WithField("entry_id", resp.EntryId).Info("CreateEntry succeeded")

	return resp, nil
}

// GetEntry handles the GetEntry RPC call
func (h *EntryHandler) GetEntry(ctx context.Context, req *bridgev1.GetEntryRequest) (*bridgev1.GetEntryResponse, error) {
	ctx, span := h.tracer.Start(ctx, "EntryHandler.GetEntry")
	defer span.End()

	h.logger.WithField("request_id", req.RequestId).Info("GetEntry RPC called")

	// Validate request
	if err := h.validateGetEntryRequest(req); err != nil {
		h.logger.WithError(err).Warn("Invalid GetEntry request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Call use case
	resp, err := h.useCase.GetEntry(ctx, req)
	if err != nil {
		h.logger.WithError(err).Error("GetEntry use case failed")
		return nil, h.mapError(err)
	}

	if !resp.Found {
		h.logger.Debug("Entry not found")
		return resp, nil
	}

	h.logger.WithField("entry_id", resp.EntryId).Info("GetEntry succeeded")

	return resp, nil
}

// UpdateEntry handles the UpdateEntry RPC call
func (h *EntryHandler) UpdateEntry(ctx context.Context, req *bridgev1.UpdateEntryRequest) (*bridgev1.UpdateEntryResponse, error) {
	ctx, span := h.tracer.Start(ctx, "EntryHandler.UpdateEntry")
	defer span.End()

	h.logger.WithFields(logrus.Fields{
		"entry_id":        req.EntryId,
		"request_id":      req.RequestId,
		"idempotency_key": req.IdempotencyKey,
	}).Info("UpdateEntry RPC called")

	// Validate request
	if err := h.validateUpdateEntryRequest(req); err != nil {
		h.logger.WithError(err).Warn("Invalid UpdateEntry request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Call use case
	resp, err := h.useCase.UpdateEntry(ctx, req)
	if err != nil {
		h.logger.WithError(err).Error("UpdateEntry use case failed")
		return nil, h.mapError(err)
	}

	h.logger.WithField("entry_id", resp.EntryId).Info("UpdateEntry succeeded")

	return resp, nil
}

// DeleteEntry handles the DeleteEntry RPC call
func (h *EntryHandler) DeleteEntry(ctx context.Context, req *bridgev1.DeleteEntryRequest) (*bridgev1.DeleteEntryResponse, error) {
	ctx, span := h.tracer.Start(ctx, "EntryHandler.DeleteEntry")
	defer span.End()

	h.logger.WithFields(logrus.Fields{
		"entry_id":        req.EntryId,
		"request_id":      req.RequestId,
		"idempotency_key": req.IdempotencyKey,
	}).Info("DeleteEntry RPC called")

	// Validate request
	if err := h.validateDeleteEntryRequest(req); err != nil {
		h.logger.WithError(err).Warn("Invalid DeleteEntry request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Call use case
	resp, err := h.useCase.DeleteEntry(ctx, req)
	if err != nil {
		h.logger.WithError(err).Error("DeleteEntry use case failed")
		return nil, h.mapError(err)
	}

	h.logger.WithField("deleted", resp.Deleted).Info("DeleteEntry succeeded")

	return resp, nil
}

// Validation methods

func (h *EntryHandler) validateCreateEntryRequest(req *bridgev1.CreateEntryRequest) error {
	if req.Key == nil {
		return status.Error(codes.InvalidArgument, "key is required")
	}
	if req.Key.KeyType == 0 {
		return status.Error(codes.InvalidArgument, "key_type is required")
	}
	if req.Key.KeyValue == "" {
		return status.Error(codes.InvalidArgument, "key_value is required")
	}
	if req.Account == nil {
		return status.Error(codes.InvalidArgument, "account is required")
	}
	if req.Account.Ispb == "" {
		return status.Error(codes.InvalidArgument, "account.ispb is required")
	}
	if req.Account.AccountNumber == "" {
		return status.Error(codes.InvalidArgument, "account.account_number is required")
	}
	if req.RequestId == "" {
		return status.Error(codes.InvalidArgument, "request_id is required")
	}
	return nil
}

func (h *EntryHandler) validateGetEntryRequest(req *bridgev1.GetEntryRequest) error {
	if req.Identifier == nil {
		return status.Error(codes.InvalidArgument, "identifier is required")
	}
	if req.RequestId == "" {
		return status.Error(codes.InvalidArgument, "request_id is required")
	}
	return nil
}

func (h *EntryHandler) validateUpdateEntryRequest(req *bridgev1.UpdateEntryRequest) error {
	if req.EntryId == "" {
		return status.Error(codes.InvalidArgument, "entry_id is required")
	}
	if req.NewAccount == nil {
		return status.Error(codes.InvalidArgument, "new_account is required")
	}
	if req.RequestId == "" {
		return status.Error(codes.InvalidArgument, "request_id is required")
	}
	return nil
}

func (h *EntryHandler) validateDeleteEntryRequest(req *bridgev1.DeleteEntryRequest) error {
	if req.EntryId == "" && req.Key == nil {
		return status.Error(codes.InvalidArgument, "entry_id or key is required")
	}
	if req.RequestId == "" {
		return status.Error(codes.InvalidArgument, "request_id is required")
	}
	return nil
}

// mapError converts domain errors to gRPC status codes
func (h *EntryHandler) mapError(err error) error {
	// Simple error mapping for now
	// TODO: Add more sophisticated error mapping based on error types

	errMsg := err.Error()

	// Map common error patterns
	switch {
	case contains(errMsg, "not found"):
		return status.Error(codes.NotFound, err.Error())
	case contains(errMsg, "already exists"):
		return status.Error(codes.AlreadyExists, err.Error())
	case contains(errMsg, "invalid"):
		return status.Error(codes.InvalidArgument, err.Error())
	case contains(errMsg, "bridge service failed"):
		return status.Error(codes.Unavailable, err.Error())
	case contains(errMsg, "database error"):
		return status.Error(codes.Internal, "internal database error")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
