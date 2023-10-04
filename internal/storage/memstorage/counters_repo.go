package memstorage

import (
	"sync"

	"github.com/Dmitrevicz/gometrics/internal/model"
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

func (s *CountersRepo) Get(name string) (model.Counter, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.counters[name]
	return v, ok
}

func (s *CountersRepo) GetAll() map[string]model.Counter {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make(map[string]model.Counter, len(s.counters))

	for k, v := range s.counters {
		res[k] = v
	}

	return res
}

func (s *CountersRepo) Set(name string, value model.Counter) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counters[name] += value
}

func (s *CountersRepo) Delete(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.counters, name)
}
