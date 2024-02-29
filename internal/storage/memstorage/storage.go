// Package memstorage implements storage repository using map as data storage.
package memstorage

import (
	"context"

	"github.com/Dmitrevicz/gometrics/internal/storage"
)

type Storage struct {
	gauges   *GaugesRepo
	counters *CountersRepo
}

func New() *Storage {
	return &Storage{}
}

func (s *Storage) Ping(ctx context.Context) error {
	return nil
}

func (s *Storage) Close(ctx context.Context) (err error) {
	return nil
}

func (s *Storage) Gauges() storage.GaugesRepository {
	if s.gauges == nil {
		s.gauges = NewGaugesRepo()
	}

	return s.gauges
}

func (s *Storage) Counters() storage.CountersRepository {
	if s.counters == nil {
		s.counters = NewCountersRepo()
	}

	return s.counters
}
