package interceptors

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	tracerName = "github.com/lbpay-lab/conn-dict/grpc"
)

// TracingInterceptor creates a gRPC unary server interceptor for OpenTelemetry distributed tracing
func TracingInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	tracer := otel.Tracer(tracerName)
	propagator := otel.GetTextMapPropagator()

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract trace context from incoming metadata
		md, _ := metadata.FromIncomingContext(ctx)
		ctx = propagator.Extract(ctx, &metadataCarrier{md: md})

		// Start a new span
		ctx, span := tracer.Start(
			ctx,
			info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("rpc.system", "grpc"),
				attribute.String("rpc.service", serviceName),
				attribute.String("rpc.method", info.FullMethod),
			),
		)
		defer span.End()

		// Add request ID to span if available
		if reqID := ctx.Value("request_id"); reqID != nil {
			if id, ok := reqID.(string); ok {
				span.SetAttributes(attribute.String("request.id", id))
			}
		}

		// Call the handler
		resp, err := handler(ctx, req)

		// Record error if present
		if err != nil {
			st, _ := status.FromError(err)
			span.SetStatus(codes.Error, st.Message())
			span.SetAttributes(
				attribute.String("rpc.grpc.status_code", st.Code().String()),
				attribute.String("error.message", st.Message()),
			)
			span.RecordError(err)
		} else {
			span.SetStatus(codes.Ok, "")
			span.SetAttributes(attribute.String("rpc.grpc.status_code", "OK"))
		}

		return resp, err
	}
}

// metadataCarrier adapts gRPC metadata to OpenTelemetry TextMapCarrier interface
type metadataCarrier struct {
	md metadata.MD
}

// Ensure metadataCarrier implements propagation.TextMapCarrier
var _ propagation.TextMapCarrier = &metadataCarrier{}

// Get retrieves a value for a key from metadata
func (mc *metadataCarrier) Get(key string) string {
	values := mc.md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// Set sets a key-value pair in metadata
func (mc *metadataCarrier) Set(key, value string) {
	mc.md.Set(key, value)
}

// Keys returns all keys in metadata
func (mc *metadataCarrier) Keys() []string {
	keys := make([]string, 0, len(mc.md))
	for k := range mc.md {
		keys = append(keys, k)
	}
	return keys
}

// TracingClientInterceptor creates a gRPC unary client interceptor for tracing outgoing requests
func TracingClientInterceptor(serviceName string) grpc.UnaryClientInterceptor {
	tracer := otel.Tracer(tracerName)
	propagator := otel.GetTextMapPropagator()

	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Start a new span for outgoing request
		ctx, span := tracer.Start(
			ctx,
			method,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(
				attribute.String("rpc.system", "grpc"),
				attribute.String("rpc.service", serviceName),
				attribute.String("rpc.method", method),
				attribute.String("peer.service", cc.Target()),
			),
		)
		defer span.End()

		// Inject trace context into outgoing metadata
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		carrier := &metadataCarrier{md: md}
		propagator.Inject(ctx, carrier)
		ctx = metadata.NewOutgoingContext(ctx, md)

		// Invoke the RPC
		err := invoker(ctx, method, req, reply, cc, opts...)

		// Record result
		if err != nil {
			st, _ := status.FromError(err)
			span.SetStatus(codes.Error, st.Message())
			span.SetAttributes(
				attribute.String("rpc.grpc.status_code", st.Code().String()),
			)
			span.RecordError(err)
		} else {
			span.SetStatus(codes.Ok, "")
			span.SetAttributes(attribute.String("rpc.grpc.status_code", "OK"))
		}

		return err
	}
}
