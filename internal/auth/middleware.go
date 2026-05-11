package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const ContextUserKey contextKey = "user_claims"

func JWTAuthMiddleware(secretKey []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "인증 토큰이 필요합니다.", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "토큰 형식이 올바르지 않습니다", http.StatusUnauthorized)
				return
			}
			tokenString := parts[1]

			claims, err := ValidateToken(tokenString, secretKey)
			if err != nil {
				http.Error(w, "유효하지 않거나 만료된 토큰입니다.", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ContextUserKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserFromContext(ctx context.Context) *CustomClaims {
	claims, ok := ctx.Value(ContextUserKey).(*CustomClaims)
	if !ok {
		return nil
	}
	return claims
}
