package storage

import "errors"

// ErrNotFound is a storage level custom error.
var ErrNotFound = errors.New("nothing found")
