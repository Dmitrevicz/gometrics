// Package storage contains storage and repository interfaces definition.
package storage

import (
	"context"

	"github.com/Dmitrevicz/gometrics/internal/model"
)

// Storage - a set of repositories
type Storage interface {
	Gauges() GaugesRepository
	Counters() CountersRepository
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
}

type GaugesRepository interface {
	// Get finds metric by name. When requested metric doesn't exist
	// storage.ErrNotFound error is returned.
	Get(name string) (model.Gauge, error)
	GetAll() (map[string]model.Gauge, error)
	Set(name string, value model.Gauge) error
	Delete(name string) error
	BatchUpdate(gauges []model.MetricGauge) (err error)
}

type CountersRepository interface {
	// Get finds metric by name. When requested metric doesn't exist
	// storage.ErrNotFound error is returned.
	Get(name string) (model.Counter, error)
	GetAll() (map[string]model.Counter, error)
	Set(name string, value model.Counter) error
	Delete(name string) error
	BatchUpdate(counters []model.MetricCounter) (err error)
}
