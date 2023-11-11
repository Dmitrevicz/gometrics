// Package postgres implements storage repository using PostgreSQL as data storage.
package postgres

import (
	"context"
	"database/sql"

	"github.com/Dmitrevicz/gometrics/internal/storage"
)

type Storage struct {
	db       *sql.DB
	gauges   *GaugesRepo
	counters *CountersRepo
}

func New(db *sql.DB) *Storage {
	return &Storage{
		db: db,
	}
}

func (s *Storage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Storage) Gauges() storage.GaugesRepository {
	if s.gauges == nil {
		s.gauges = NewGaugesRepo(s)
	}

	return s.gauges
}

func (s *Storage) Counters() storage.CountersRepository {
	if s.counters == nil {
		s.counters = NewCountersRepo(s)
	}

	return s.counters
}
