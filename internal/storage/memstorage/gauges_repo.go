package memstorage

import (
	"sync"

	"github.com/Dmitrevicz/gometrics/internal/model"
)

type GaugesRepo struct {
	mu     sync.RWMutex
	gauges map[string]model.Gauge
}

func NewGaugesRepo() *GaugesRepo {
	return &GaugesRepo{
		gauges: make(map[string]model.Gauge),
	}
}

func (s *GaugesRepo) Get(name string) (model.Gauge, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.gauges[name]
	return v, ok
}

func (s *GaugesRepo) GetAll() map[string]model.Gauge {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make(map[string]model.Gauge, len(s.gauges))

	for k, v := range s.gauges {
		res[k] = v
	}

	return res
}

func (s *GaugesRepo) Set(name string, value model.Gauge) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.gauges[name] = value
}

func (s *GaugesRepo) Delete(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.gauges, name)
}
