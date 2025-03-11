package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"toptal/internal/app/domain"
	"toptal/internal/app/handler/model"
	"toptal/internal/app/util"
)

type AuthService interface {
	GetUserById(ctx context.Context, id int) (domain.User, error)
}

// RoleMiddleware returns a middleware function that checks if the user has the required role.
// Currently it supports only the \"admin\" role by calling the AuthService.checkAdmin method.
func RoleMiddleware(authService AuthService, next func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, err := util.GetUserID(r.Context())
		if err != nil {
			slog.Error("failed to get user ID from context", "error", err)
			model.Unauthorized(w, "failed to get user ID from context", r.URL.Path)
			return
		}
		user, err := authService.GetUserById(r.Context(), userId)
		if err != nil {
			slog.Error("failed to find user by ID", "userId", userId, "error", err)
			model.Unauthorized(w, "failed to find user by ID", r.URL.Path)
			return
		}
		if !user.Admin() {
			model.Forbidden(w, "user does not have admin role", r.URL.Path)
			return
		}

		next(w, r)
	}
}
