package model

import (
	"errors"
	"fmt"
)

var ErrDatabase = errors.New("database error")

func WrapDatabaseError(err error, msg string) error {
	return fmt.Errorf("%s: %w: %v", msg, ErrDatabase, err)
}
