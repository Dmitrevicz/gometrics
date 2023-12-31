package memstorage

import (
	"sync"

	"github.com/Dmitrevicz/gometrics/internal/model"
	"github.com/Dmitrevicz/gometrics/internal/storage"
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

// Get finds metric by name. When requested metric doesn't exist
// storage.ErrNotFound error is returned.
func (s *GaugesRepo) Get(name string) (model.Gauge, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.gauges[name]
	if !ok {
		return 0, storage.ErrNotFound
	}

	return v, nil
}

func (s *GaugesRepo) GetAll() (map[string]model.Gauge, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make(map[string]model.Gauge, len(s.gauges))

	for k, v := range s.gauges {
		res[k] = v
	}

	return res, nil
}

func (s *GaugesRepo) Set(name string, value model.Gauge) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.gauges[name] = value
	return nil
}

func (s *GaugesRepo) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.gauges, name)
	return nil
}

func (s *GaugesRepo) BatchUpdate(gauges []model.MetricGauge) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, g := range gauges {
		s.gauges[g.Name] = g.Value
	}

	return nil
}
