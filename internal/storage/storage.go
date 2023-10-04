package storage

import (
	"sync"

	"github.com/Dmitrevicz/gometrics/internal/model"
)

// TODO: refactor with interfaces later
type Storage struct {
	Gauges   *GaugesMap
	Counters *CountersMap
}

func New() *Storage {
	return &Storage{
		Gauges:   NewGaugesMap(),
		Counters: NewCountersMap(),
	}
}

type GaugesMap struct {
	m      sync.RWMutex
	gauges map[string]model.Gauge
}

func NewGaugesMap() *GaugesMap {
	return &GaugesMap{
		gauges: make(map[string]model.Gauge),
	}
}

func (s *GaugesMap) Get(name string) (model.Gauge, bool) {
	s.m.RLock()
	defer s.m.RUnlock()

	v, ok := s.gauges[name]
	return v, ok
}

func (s *GaugesMap) GetAll() map[string]model.Gauge {
	s.m.RLock()
	defer s.m.RUnlock()

	res := make(map[string]model.Gauge, len(s.gauges))

	for k, v := range s.gauges {
		res[k] = v
	}

	return res
}

func (s *GaugesMap) Set(name string, value model.Gauge) {
	s.m.Lock()
	defer s.m.Unlock()

	s.gauges[name] = value
}

func (s *GaugesMap) Remove(name string) {
	s.m.Lock()
	defer s.m.Unlock()

	delete(s.gauges, name)
}

type CountersMap struct {
	m        sync.RWMutex
	counters map[string]model.Counter
}

func NewCountersMap() *CountersMap {
	return &CountersMap{
		counters: make(map[string]model.Counter),
	}
}

func (s *CountersMap) Get(name string) (model.Counter, bool) {
	s.m.RLock()
	defer s.m.RUnlock()

	v, ok := s.counters[name]
	return v, ok
}

func (s *CountersMap) GetAll() map[string]model.Counter {
	s.m.RLock()
	defer s.m.RUnlock()

	res := make(map[string]model.Counter, len(s.counters))

	for k, v := range s.counters {
		res[k] = v
	}

	return res
}

func (s *CountersMap) Set(name string, value model.Counter) {
	s.m.Lock()
	defer s.m.Unlock()

	s.counters[name] += value
}

func (s *CountersMap) Remove(name string) {
	s.m.Lock()
	defer s.m.Unlock()

	delete(s.counters, name)
}
