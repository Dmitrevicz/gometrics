package model

import "fmt"

type RetriableError error

func NewRetriableError(err error) error {
	return fmt.Errorf("%w", err)
}
