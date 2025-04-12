package middleware

import (
	"avito-intern/internal/models"
	"avito-intern/internal/utils"
	"context"
	"errors"
	"net/http"
	"strings"
)

type contextKey string

const UserCtxKey = contextKey("user")

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Invalid Authorization header", http.StatusUnauthorized)
			return
		}

		claims, err := utils.ParseJWT(parts[1])
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		user := models.User{
			ID:    claims["user_id"].(string),
			Role:  claims["role"].(string),
			Email: "",
		}

		ctx := context.WithValue(r.Context(), UserCtxKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(ctx context.Context) (models.User, error) {
	user, ok := ctx.Value(UserCtxKey).(models.User)
	if !ok {
		return models.User{}, errors.New("user not found in context")
	}
	return user, nil
}

func RequireRole(ctx context.Context, required string) error {
	user, err := GetUserFromContext(ctx)
	if err != nil {
		return err
	}
	if user.Role != required {
		return errors.New("access denied")
	}
	return nil
}
