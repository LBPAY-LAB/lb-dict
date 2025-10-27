package handlers

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InfractionService interface for business logic
// This interface defines the contract between the handler and the service layer
type InfractionService interface {
	CreateInfraction(ctx context.Context, req interface{}) (interface{}, error)
	InvestigateInfraction(ctx context.Context, req interface{}) (interface{}, error)
	ResolveInfraction(ctx context.Context, req interface{}) (interface{}, error)
	DismissInfraction(ctx context.Context, req interface{}) (interface{}, error)
	GetInfraction(ctx context.Context, req interface{}) (interface{}, error)
	ListInfractions(ctx context.Context, req interface{}) (interface{}, error)
}

// InfractionHandler implements the gRPC handlers for Infraction operations
// This handler is responsible for:
// - Input validation
// - Request logging and tracing
// - Delegating business logic to InfractionService
// - Error mapping from domain to gRPC errors
type InfractionHandler struct {
	// Note: UnimplementedInfractionServiceServer will be added when proto is generated
	// pb.UnimplementedInfractionServiceServer

	service InfractionService
	logger  *logrus.Logger
	tracer  trace.Tracer
}

// NewInfractionHandler creates a new InfractionHandler instance
func NewInfractionHandler(service InfractionService, logger *logrus.Logger, tracer trace.Tracer) *InfractionHandler {
	return &InfractionHandler{
		service: service,
		logger:  logger,
		tracer:  tracer,
	}
}

// CreateInfraction handles the CreateInfraction RPC call
// Initiates an infraction investigation workflow via Temporal
func (h *InfractionHandler) CreateInfraction(ctx context.Context, req interface{}) (interface{}, error) {
	ctx, span := h.tracer.Start(ctx, "InfractionHandler.CreateInfraction")
	defer span.End()

	h.logger.Info("CreateInfraction RPC called")

	// Validate request format
	if err := h.validateRequest(req); err != nil {
		h.logger.WithError(err).Warn("Invalid CreateInfraction request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Call service layer
	resp, err := h.service.CreateInfraction(ctx, req)
	if err != nil {
		h.logger.WithError(err).Error("CreateInfraction service failed")
		return nil, h.mapError(err)
	}

	h.logger.Info("CreateInfraction succeeded")
	return resp, nil
}

// InvestigateInfraction handles the InvestigateInfraction RPC call
// Sends investigation decision signal to the infraction workflow
func (h *InfractionHandler) InvestigateInfraction(ctx context.Context, req interface{}) (interface{}, error) {
	ctx, span := h.tracer.Start(ctx, "InfractionHandler.InvestigateInfraction")
	defer span.End()

	h.logger.Info("InvestigateInfraction RPC called")

	// Validate request format
	if err := h.validateRequest(req); err != nil {
		h.logger.WithError(err).Warn("Invalid InvestigateInfraction request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Call service layer
	resp, err := h.service.InvestigateInfraction(ctx, req)
	if err != nil {
		h.logger.WithError(err).Error("InvestigateInfraction service failed")
		return nil, h.mapError(err)
	}

	h.logger.Info("InvestigateInfraction succeeded")
	return resp, nil
}

// ResolveInfraction handles the ResolveInfraction RPC call
// Convenience method that sends a RESOLVE decision to the workflow
func (h *InfractionHandler) ResolveInfraction(ctx context.Context, req interface{}) (interface{}, error) {
	ctx, span := h.tracer.Start(ctx, "InfractionHandler.ResolveInfraction")
	defer span.End()

	h.logger.Info("ResolveInfraction RPC called")

	// Validate request format
	if err := h.validateRequest(req); err != nil {
		h.logger.WithError(err).Warn("Invalid ResolveInfraction request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Call service layer
	resp, err := h.service.ResolveInfraction(ctx, req)
	if err != nil {
		h.logger.WithError(err).Error("ResolveInfraction service failed")
		return nil, h.mapError(err)
	}

	h.logger.Info("ResolveInfraction succeeded")
	return resp, nil
}

// DismissInfraction handles the DismissInfraction RPC call
// Convenience method that sends a DISMISS decision to the workflow
func (h *InfractionHandler) DismissInfraction(ctx context.Context, req interface{}) (interface{}, error) {
	ctx, span := h.tracer.Start(ctx, "InfractionHandler.DismissInfraction")
	defer span.End()

	h.logger.Info("DismissInfraction RPC called")

	// Validate request format
	if err := h.validateRequest(req); err != nil {
		h.logger.WithError(err).Warn("Invalid DismissInfraction request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Call service layer
	resp, err := h.service.DismissInfraction(ctx, req)
	if err != nil {
		h.logger.WithError(err).Error("DismissInfraction service failed")
		return nil, h.mapError(err)
	}

	h.logger.Info("DismissInfraction succeeded")
	return resp, nil
}

// GetInfraction handles the GetInfraction RPC call
// Retrieves a single infraction by ID from the database
func (h *InfractionHandler) GetInfraction(ctx context.Context, req interface{}) (interface{}, error) {
	ctx, span := h.tracer.Start(ctx, "InfractionHandler.GetInfraction")
	defer span.End()

	h.logger.Info("GetInfraction RPC called")

	// Validate request format
	if err := h.validateRequest(req); err != nil {
		h.logger.WithError(err).Warn("Invalid GetInfraction request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Call service layer
	resp, err := h.service.GetInfraction(ctx, req)
	if err != nil {
		h.logger.WithError(err).Error("GetInfraction service failed")
		return nil, h.mapError(err)
	}

	h.logger.Info("GetInfraction succeeded")
	return resp, nil
}

// ListInfractions handles the ListInfractions RPC call
// Retrieves a paginated list of infractions from the database
func (h *InfractionHandler) ListInfractions(ctx context.Context, req interface{}) (interface{}, error) {
	ctx, span := h.tracer.Start(ctx, "InfractionHandler.ListInfractions")
	defer span.End()

	h.logger.Info("ListInfractions RPC called")

	// Validate request format
	if err := h.validateRequest(req); err != nil {
		h.logger.WithError(err).Warn("Invalid ListInfractions request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Call service layer
	resp, err := h.service.ListInfractions(ctx, req)
	if err != nil {
		h.logger.WithError(err).Error("ListInfractions service failed")
		return nil, h.mapError(err)
	}

	h.logger.Info("ListInfractions succeeded")
	return resp, nil
}

// Validation methods

// validateRequest performs basic request validation
func (h *InfractionHandler) validateRequest(req interface{}) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request is required")
	}

	// Additional validation will be done at the service layer
	// where we have access to strongly-typed fields
	return nil
}

// mapError converts domain/service errors to gRPC status codes
func (h *InfractionHandler) mapError(err error) error {
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
	case contains(errMsg, "workflow"):
		return status.Error(codes.Internal, "workflow execution error")
	case contains(errMsg, "database"):
		return status.Error(codes.Internal, "internal database error")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
