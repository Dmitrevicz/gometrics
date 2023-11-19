package memstorage

import (
	"sync"

	"github.com/Dmitrevicz/gometrics/internal/model"
	"github.com/Dmitrevicz/gometrics/internal/storage"
)

type CountersRepo struct {
	mu       sync.RWMutex
	counters map[string]model.Counter
}

func NewCountersRepo() *CountersRepo {
	return &CountersRepo{
		counters: make(map[string]model.Counter),
	}
}

// Get finds metric by name. When requested metric doesn't exist
// storage.ErrNotFound error is returned.
func (s *CountersRepo) Get(name string) (model.Counter, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.counters[name]
	if !ok {
		return 0, storage.ErrNotFound
	}

	return v, nil
}

func (s *CountersRepo) GetAll() (map[string]model.Counter, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make(map[string]model.Counter, len(s.counters))

	for k, v := range s.counters {
		res[k] = v
	}

	return res, nil
}

func (s *CountersRepo) Set(name string, value model.Counter) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counters[name] += value
	return nil
}

func (s *CountersRepo) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.counters, name)
	return nil
}

func (s *CountersRepo) BatchUpdate(counters []model.MetricCounter) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, c := range counters {
		s.counters[c.Name] += c.Value
	}

	return nil
}
