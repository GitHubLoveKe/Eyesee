package middleware

import (
	"context"
	"dcs/internal/config"
	"dcs/internal/service"
	"dcs/pkg/response"
	"net/http"
	"strings"
)

type contextKey string

const claimsKey contextKey = "jwt_claims"

func JWT(authSvc *service.AuthService, cfg *config.JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Unauthorized(w, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				response.Unauthorized(w, "invalid authorization header format")
				return
			}

			tokenStr := parts[1]
			claims, err := authSvc.ValidateToken(tokenStr)
			if err != nil {
				response.Unauthorized(w, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) string {
	claims, ok := ctx.Value(claimsKey).(*service.Claims)
	if !ok {
		return ""
	}
	return claims.UserID
}

func UsernameFromContext(ctx context.Context) string {
	claims, ok := ctx.Value(claimsKey).(*service.Claims)
	if !ok {
		return ""
	}
	return claims.Username
}

func UserRoleFromContext(ctx context.Context) string {
	claims, ok := ctx.Value(claimsKey).(*service.Claims)
	if !ok {
		return ""
	}
	return claims.Role
}
