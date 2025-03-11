package util

import (
	"context"
	"fmt"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
)

func GetUserID(ctx context.Context) (int, error) {
	val := ctx.Value(UserIDKey)
	if val == nil {
		return 0, fmt.Errorf("user ID not found in context")
	}

	userID, ok := val.(int)
	if !ok {
		return 0, fmt.Errorf("invalid user ID type in context")
	}

	return userID, nil
}

func WithUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}
