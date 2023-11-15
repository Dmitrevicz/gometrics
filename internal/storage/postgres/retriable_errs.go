package postgres

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// CheckRetriableErrors return value shows whether action that lead to err
// should be retried.
func CheckRetriableErrors(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgerrcode.IsConnectionException(pgErr.Code) {
			return true
		}
	}

	return false
}
