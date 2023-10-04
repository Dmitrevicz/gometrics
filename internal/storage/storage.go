package storage

import "github.com/Dmitrevicz/gometrics/internal/model"

// Storage - a set of repositories
type Storage interface {
	Gauges() GaugesRepository
	Counters() CountersRepository
}

type GaugesRepository interface {
	Get(name string) (model.Gauge, bool)
	GetAll() map[string]model.Gauge
	Set(name string, value model.Gauge)
	Delete(name string)
}

type CountersRepository interface {
	Get(name string) (model.Counter, bool)
	GetAll() map[string]model.Counter
	Set(name string, value model.Counter)
	Delete(name string)
}
