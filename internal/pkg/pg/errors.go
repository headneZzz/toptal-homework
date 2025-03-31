package pg

import (
	"errors"
	"github.com/lib/pq"
)

const (
	uniqueViolationErr     = "23505"
	foreignKeyViolationErr = "23503"
)

func IsUniqueViolationErr(err error) bool {
	var pqErr *pq.Error
	if ok := errors.As(err, &pqErr); ok && pqErr.Code == uniqueViolationErr {
		return true
	}
	return false
}

func IsForeignKeyViolationErr(err error) bool {
	var pqErr *pq.Error
	if ok := errors.As(err, &pqErr); ok && pqErr.Code == foreignKeyViolationErr {
		return true
	}
	return false
}
