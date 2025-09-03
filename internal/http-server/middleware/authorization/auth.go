package authorization

import (
	"context"
	"log/slog"
	"money-tracker/internal/lib/jwt"
	"money-tracker/internal/lib/logger/sl"
	"net/http"
	"strings"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func Auth(log *slog.Logger, jwtSecret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log := log.With(
			slog.String("component", "middleware/auth"),
		)

		log.Info("auth middleware enabled")

		fn := func(w http.ResponseWriter, r *http.Request) {
			const op = "middleware.auth.Auth"

			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
                http.Error(w, "Authorization header required", http.StatusUnauthorized)
                return
            }

            parts := strings.Split(authHeader, " ")
            if len(parts) != 2 || parts[0] != "Bearer" {
                http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
                return
            }

            token := parts[1]
            claims, err := jwt.ParseToken(token, jwtSecret)
            if err != nil {
                log.Error("failed to parse token", sl.Err(err), slog.String("op", op))
                http.Error(w, "Invalid token", http.StatusUnauthorized)
                return
            }

            ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
            next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}