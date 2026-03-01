package util

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthInterceptor validates the JWT from the Authorization header and injects claims into the context
func AuthInterceptor(jwtSecret string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.New("missing metadata")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			// Allow unauthenticated requests to proceed, individual services can check for claims if needed
			return handler(ctx, req)
		}

		tokenString := strings.TrimPrefix(authHeader[0], "Bearer ")
		claims, err := ValidateToken(tokenString, jwtSecret)
		if err != nil {
			return nil, errors.New("invalid or expired token")
		}

		newCtx := context.WithValue(ctx, AccountIDKey, claims.AccountID)
		newCtx = context.WithValue(newCtx, UserTypeKey, claims.UserType)
		newCtx = context.WithValue(newCtx, EmailKey, claims.Email)

		return handler(newCtx, req)
	}
}
