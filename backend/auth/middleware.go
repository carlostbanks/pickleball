// pickle/backend/auth/middleware.go
package auth

import (
	"context"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// contextKey is a custom type for context keys
type contextKey string

// UserIDKey is the context key for user ID
const UserIDKey contextKey = "user_id"

// AuthMiddleware is a middleware for HTTP endpoints that require authentication
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from cookie or Authorization header
		var tokenString string

		// Try to get token from cookie first
		cookie, err := r.Cookie("token")
		if err == nil {
			tokenString = cookie.Value
		}

		// If no token in cookie, try Authorization header
		if tokenString == "" {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				// Check if it's a Bearer token
				if strings.HasPrefix(authHeader, "Bearer ") {
					tokenString = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}
		}

		// If no token found, return unauthorized
		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := ValidateJWT(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Create a new context with the user ID
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)

		// Call the next handler with the new context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID retrieves the user ID from the context
func GetUserID(ctx context.Context) string {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return ""
	}
	return userID
}

// AuthInterceptor is a gRPC interceptor for authentication
func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Skip authentication for public methods
	if isPublicMethod(info.FullMethod) {
		return handler(ctx, req)
	}

	// Get token from metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	// Get authorization header
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	// Get token from authorization header
	authHeader := authHeaders[0]
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, status.Errorf(codes.Unauthenticated, "invalid authorization token format")
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Validate token
	claims, err := ValidateJWT(tokenString)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	// Create a new context with the user ID
	newCtx := context.WithValue(ctx, UserIDKey, claims.UserID)

	// Call the handler with the new context
	return handler(newCtx, req)
}

// isPublicMethod checks if the method is public (doesn't require authentication)
func isPublicMethod(method string) bool {
	// List of public methods (doesn't require authentication)
	publicMethods := []string{
		"/scheduler.SchedulerService/GetCourts",
		"/scheduler.SchedulerService/GetCourt",
	}

	for _, publicMethod := range publicMethods {
		if publicMethod == method {
			return true
		}
	}

	return false
}
