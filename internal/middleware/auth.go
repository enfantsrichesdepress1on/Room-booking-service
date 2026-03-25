package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"room-booking-service/internal/auth"
)

type contextKey string

const ClaimsContextKey contextKey = "claims"

func Auth(manager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				writeErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing bearer token")
				return
			}
			claims, err := manager.Parse(strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
			if err != nil {
				writeErrorJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token")
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ClaimsContextKey, claims)))
		})
	}
}

func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r.Context())
			if claims == nil || claims.Role != role {
				writeErrorJSON(w, http.StatusForbidden, "FORBIDDEN", "forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func GetClaims(ctx context.Context) *auth.Claims {
	claims, _ := ctx.Value(ClaimsContextKey).(*auth.Claims)
	return claims
}

func writeErrorJSON(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]string{"code": code, "message": message}})
}
