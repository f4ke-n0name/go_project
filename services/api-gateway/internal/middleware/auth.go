package middleware

import (
	"context"
	"net/http"
	"strings"

	auth "github.com/f4ke-n0name/autoparts-hub/gen/auth"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
)

func Auth(authClient auth.AuthClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			resp, err := authClient.ValidateToken(r.Context(), &auth.ValidateRequest{
				AccessToken: token,
			})
			if err != nil || !resp.GetValid() {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, resp.GetUserId())
			ctx = context.WithValue(ctx, RoleKey, resp.GetRole())

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request) string {
	bearer := r.Header.Get("Authorization")
	if bearer == "" {
		return ""
	}
	parts := strings.SplitN(bearer, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}

func GetUserID(ctx context.Context) string {
	v, _ := ctx.Value(UserIDKey).(string)
	return v
}

func GetRole(ctx context.Context) string {
	v, _ := ctx.Value(RoleKey).(string)
	return v
}
