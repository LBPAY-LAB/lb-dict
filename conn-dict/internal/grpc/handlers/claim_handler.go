package handlers

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ClaimService interface for business logic
// This interface defines the contract between the handler and the service layer
type ClaimService interface {
	CreateClaim(ctx context.Context, req interface{}) (interface{}, error)
	ConfirmClaim(ctx context.Context, req interface{}) (interface{}, error)
	CancelClaim(ctx context.Context, req interface{}) (interface{}, error)
	GetClaim(ctx context.Context, req interface{}) (interface{}, error)
	ListClaims(ctx context.Context, req interface{}) (interface{}, error)
}

// ClaimHandler implements the gRPC handlers for Claim operations
// This handler is responsible for:
// - Input validation
// - Request logging and tracing
// - Delegating business logic to ClaimService
// - Error mapping from domain to gRPC errors
type ClaimHandler struct {
	// Note: UnimplementedClaimServiceServer will be added when proto is generated
	// pb.UnimplementedClaimServiceServer

	service ClaimService
	logger  *logrus.Logger
	tracer  trace.Tracer
}

// NewClaimHandler creates a new ClaimHandler instance
func NewClaimHandler(service ClaimService, logger *logrus.Logger, tracer trace.Tracer) *ClaimHandler {
	return &ClaimHandler{
		service: service,
		logger:  logger,
		tracer:  tracer,
	}
}

// CreateClaim handles the CreateClaim RPC call
// Initiates a 30-day claim workflow via Temporal
func (h *ClaimHandler) CreateClaim(ctx context.Context, req interface{}) (interface{}, error) {
	ctx, span := h.tracer.Start(ctx, "ClaimHandler.CreateClaim")
	defer span.End()

	h.logger.Info("CreateClaim RPC called")

	// Validate request format
	if err := h.validateRequest(req); err != nil {
		h.logger.WithError(err).Warn("Invalid CreateClaim request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Call service layer
	resp, err := h.service.CreateClaim(ctx, req)
	if err != nil {
		h.logger.WithError(err).Error("CreateClaim service failed")
		return nil, h.mapError(err)
	}

	h.logger.Info("CreateClaim succeeded")
	return resp, nil
}

// ConfirmClaim handles the ConfirmClaim RPC call
// Sends confirmation signal to the claim workflow
func (h *ClaimHandler) ConfirmClaim(ctx context.Context, req interface{}) (interface{}, error) {
	ctx, span := h.tracer.Start(ctx, "ClaimHandler.ConfirmClaim")
	defer span.End()

	h.logger.Info("ConfirmClaim RPC called")

	// Validate request format
	if err := h.validateRequest(req); err != nil {
		h.logger.WithError(err).Warn("Invalid ConfirmClaim request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Call service layer
	resp, err := h.service.ConfirmClaim(ctx, req)
	if err != nil {
		h.logger.WithError(err).Error("ConfirmClaim service failed")
		return nil, h.mapError(err)
	}

	h.logger.Info("ConfirmClaim succeeded")
	return resp, nil
}

// CancelClaim handles the CancelClaim RPC call
// Sends cancellation signal to the claim workflow
func (h *ClaimHandler) CancelClaim(ctx context.Context, req interface{}) (interface{}, error) {
	ctx, span := h.tracer.Start(ctx, "ClaimHandler.CancelClaim")
	defer span.End()

	h.logger.Info("CancelClaim RPC called")

	// Validate request format
	if err := h.validateRequest(req); err != nil {
		h.logger.WithError(err).Warn("Invalid CancelClaim request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Call service layer
	resp, err := h.service.CancelClaim(ctx, req)
	if err != nil {
		h.logger.WithError(err).Error("CancelClaim service failed")
		return nil, h.mapError(err)
	}

	h.logger.Info("CancelClaim succeeded")
	return resp, nil
}

// GetClaim handles the GetClaim RPC call
// Retrieves a single claim by ID from the database
func (h *ClaimHandler) GetClaim(ctx context.Context, req interface{}) (interface{}, error) {
	ctx, span := h.tracer.Start(ctx, "ClaimHandler.GetClaim")
	defer span.End()

	h.logger.Info("GetClaim RPC called")

	// Validate request format
	if err := h.validateRequest(req); err != nil {
		h.logger.WithError(err).Warn("Invalid GetClaim request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Call service layer
	resp, err := h.service.GetClaim(ctx, req)
	if err != nil {
		h.logger.WithError(err).Error("GetClaim service failed")
		return nil, h.mapError(err)
	}

	h.logger.Info("GetClaim succeeded")
	return resp, nil
}

// ListClaims handles the ListClaims RPC call
// Retrieves a paginated list of claims from the database
func (h *ClaimHandler) ListClaims(ctx context.Context, req interface{}) (interface{}, error) {
	ctx, span := h.tracer.Start(ctx, "ClaimHandler.ListClaims")
	defer span.End()

	h.logger.Info("ListClaims RPC called")

	// Validate request format
	if err := h.validateRequest(req); err != nil {
		h.logger.WithError(err).Warn("Invalid ListClaims request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Call service layer
	resp, err := h.service.ListClaims(ctx, req)
	if err != nil {
		h.logger.WithError(err).Error("ListClaims service failed")
		return nil, h.mapError(err)
	}

	h.logger.Info("ListClaims succeeded")
	return resp, nil
}

// Validation methods

// validateRequest performs basic request validation
func (h *ClaimHandler) validateRequest(req interface{}) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request is required")
	}

	// Additional validation will be done at the service layer
	// where we have access to strongly-typed fields
	return nil
}

// mapError converts domain/service errors to gRPC status codes
func (h *ClaimHandler) mapError(err error) error {
	// If already a gRPC status, return as-is
	if _, ok := status.FromError(err); ok {
		return err
	}

	// Map common error patterns
	errMsg := err.Error()

	switch {
	case contains(errMsg, "not found"):
		return status.Error(codes.NotFound, err.Error())
	case contains(errMsg, "already exists"):
		return status.Error(codes.AlreadyExists, err.Error())
	case contains(errMsg, "invalid"):
		return status.Error(codes.InvalidArgument, err.Error())
	case contains(errMsg, "expired"):
		return status.Error(codes.FailedPrecondition, err.Error())
	case contains(errMsg, "cannot be cancelled"):
		return status.Error(codes.FailedPrecondition, err.Error())
	case contains(errMsg, "workflow"):
		return status.Error(codes.Internal, "workflow execution error")
	case contains(errMsg, "database"):
		return status.Error(codes.Internal, "internal database error")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
