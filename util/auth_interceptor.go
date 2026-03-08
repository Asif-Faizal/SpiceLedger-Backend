package util

import (
	"context"
	"encoding/base64"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthInterceptor validates the JWT from the Authorization header and injects claims into the context
func AuthInterceptor(jwtSecret, basicUser, basicPass string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		newCtx := context.WithValue(ctx, IsAuthenticatedKey, false)
		newCtx = context.WithValue(newCtx, IsAdminKey, false)

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(newCtx, req)
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return handler(newCtx, req)
		}

		headerValue := authHeader[0]

		if strings.HasPrefix(headerValue, "Bearer ") {
			tokenString := strings.TrimPrefix(headerValue, "Bearer ")
			claims, err := ValidateToken(tokenString, jwtSecret)
			if err == nil {
				newCtx = context.WithValue(newCtx, AccountIDKey, claims.AccountID)
				newCtx = context.WithValue(newCtx, UserTypeKey, claims.UserType)
				newCtx = context.WithValue(newCtx, EmailKey, claims.Email)
				newCtx = context.WithValue(newCtx, IsAuthenticatedKey, true)
				newCtx = context.WithValue(newCtx, AccessTokenKey, tokenString)
				if claims.UserType == UserTypeAdmin {
					newCtx = context.WithValue(newCtx, IsAdminKey, true)
				}
			}
			return handler(newCtx, req)
		}

		if strings.HasPrefix(headerValue, "Basic ") {
			encoded := strings.TrimPrefix(headerValue, "Basic ")
			decoded, err := base64.StdEncoding.DecodeString(encoded)
			if err == nil {
				parts := strings.SplitN(string(decoded), ":", 2)
				if len(parts) == 2 && parts[0] == basicUser && parts[1] == basicPass {
					newCtx = context.WithValue(newCtx, IsAuthenticatedKey, true)
				}
			}
		}

		return handler(newCtx, req)
	}
}
