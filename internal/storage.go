package internal

import "sync"

type (
	gauge   float64
	counter int64
)

// TODO: refactor with interfaces later
type Storage struct {
	Gauges   *GaugesMap
	Counters *CountersMap
}

func NewStorage() *Storage {
	return &Storage{
		Gauges:   NewGaugesMap(),
		Counters: NewCountersMap(),
	}
}

type GaugesMap struct {
	m      sync.RWMutex
	gauges map[string]gauge
}

func NewGaugesMap() *GaugesMap {
	return &GaugesMap{
		gauges: make(map[string]gauge),
	}
}

func (s *GaugesMap) Get(name string) (gauge, bool) {
	s.m.RLock()
	defer s.m.RUnlock()

	v, ok := s.gauges[name]
	return v, ok
}

func (s *GaugesMap) GetAll() map[string]gauge {
	s.m.RLock()
	defer s.m.RUnlock()

	res := make(map[string]gauge, len(s.gauges))

	for k, v := range s.gauges {
		res[k] = v
	}

	return res
}

func (s *GaugesMap) Set(name string, value gauge) {
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
	counters map[string]counter
}

func NewCountersMap() *CountersMap {
	return &CountersMap{
		counters: make(map[string]counter),
	}
}

func (s *CountersMap) Get(name string) (counter, bool) {
	s.m.RLock()
	defer s.m.RUnlock()

	v, ok := s.counters[name]
	return v, ok
}

func (s *CountersMap) GetAll() map[string]counter {
	s.m.RLock()
	defer s.m.RUnlock()

	res := make(map[string]counter, len(s.counters))

	for k, v := range s.counters {
		res[k] = v
	}

	return res
}

func (s *CountersMap) Set(name string, value counter) {
	s.m.Lock()
	defer s.m.Unlock()

	s.counters[name] += value
}

func (s *CountersMap) Remove(name string) {
	s.m.Lock()
	defer s.m.Unlock()

	delete(s.counters, name)
}
