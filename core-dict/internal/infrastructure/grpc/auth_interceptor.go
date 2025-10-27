package grpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor handles JWT authentication for gRPC requests
type AuthInterceptor struct {
	jwtSecret      string
	skipAuthMethods map[string]bool
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret       string
	SkipAuthMethods []string // Methods that don't require authentication (e.g., HealthCheck)
}

// NewAuthInterceptor creates a new authentication interceptor
func NewAuthInterceptor(config *AuthConfig) *AuthInterceptor {
	if config == nil {
		config = &AuthConfig{
			JWTSecret: "default-secret-change-me",
		}
	}

	skipAuth := make(map[string]bool)
	for _, method := range config.SkipAuthMethods {
		skipAuth[method] = true
	}

	// Always skip auth for HealthCheck
	skipAuth["/dict.core.v1.CoreDictService/HealthCheck"] = true

	return &AuthInterceptor{
		jwtSecret:       config.JWTSecret,
		skipAuthMethods: skipAuth,
	}
}

// Unary returns a unary server interceptor for authentication
func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip authentication for certain methods
		if i.skipAuthMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// Extract metadata from context
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		// Get authorization header
		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		// Extract token from "Bearer <token>"
		authHeader := authHeaders[0]
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			return nil, status.Error(codes.Unauthenticated, "empty token")
		}

		// Validate JWT token
		claims, err := i.validateJWT(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("invalid token: %v", err))
		}

		// Add user information to context for downstream handlers
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_role", claims.Role)
		ctx = context.WithValue(ctx, "ispb", claims.ISPB)

		// Call the handler
		return handler(ctx, req)
	}
}

// JWTClaims represents the claims in a JWT token
type JWTClaims struct {
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`      // user, admin, support
	ISPB      string    `json:"ispb"`      // Bank ISPB
	ExpiresAt time.Time `json:"exp"`
	IssuedAt  time.Time `json:"iat"`
}

// validateJWT validates a JWT token and returns the claims
// TODO: Replace with proper JWT library (e.g., github.com/golang-jwt/jwt)
func (i *AuthInterceptor) validateJWT(token string) (*JWTClaims, error) {
	// This is a simplified mock implementation
	// In production, use a proper JWT library to:
	// 1. Verify signature with i.jwtSecret
	// 2. Check expiration
	// 3. Validate issuer, audience, etc.

	// For now, accept any token and extract mock claims
	// TODO: Implement proper JWT validation
	if len(token) < 10 {
		return nil, fmt.Errorf("token too short")
	}

	// Mock claims for development
	return &JWTClaims{
		UserID:    "user-123",
		Role:      "user",
		ISPB:      "12345678",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IssuedAt:  time.Now(),
	}, nil
}

// Helper functions to extract auth context from downstream handlers

// GetUserID extracts the user ID from the context
func GetUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return "", fmt.Errorf("user_id not found in context")
	}
	return userID, nil
}

// GetUserRole extracts the user role from the context
func GetUserRole(ctx context.Context) (string, error) {
	role, ok := ctx.Value("user_role").(string)
	if !ok {
		return "", fmt.Errorf("user_role not found in context")
	}
	return role, nil
}

// GetISPB extracts the ISPB from the context
func GetISPB(ctx context.Context) (string, error) {
	ispb, ok := ctx.Value("ispb").(string)
	if !ok {
		return "", fmt.Errorf("ispb not found in context")
	}
	return ispb, nil
}

// CheckPermission checks if the user has the required role
func CheckPermission(ctx context.Context, requiredRoles ...string) error {
	role, err := GetUserRole(ctx)
	if err != nil {
		return status.Error(codes.PermissionDenied, "user role not found")
	}

	for _, required := range requiredRoles {
		if role == required {
			return nil
		}
	}

	return status.Errorf(codes.PermissionDenied, "insufficient permissions (required: %v, have: %s)", requiredRoles, role)
}
