package model

import "fmt"

// RetriableError is a custom error type that should be retried.
type RetriableError error

// NewRetriableError creates new RetriableError that wraps err.
func NewRetriableError(err error) error {
	return fmt.Errorf("%w", err)
}
