package memstorage

import "github.com/Dmitrevicz/gometrics/internal/storage"

type Storage struct {
	gauges   *GaugesRepo
	counters *CountersRepo
}

func New() *Storage {
	return &Storage{}
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
