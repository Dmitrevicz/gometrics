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
}

type GaugesRepository interface {
	Get(name string) (model.Gauge, bool, error)
	GetAll() (map[string]model.Gauge, error)
	Set(name string, value model.Gauge) error
	Delete(name string) error
}

type CountersRepository interface {
	Get(name string) (model.Counter, bool, error)
	GetAll() (map[string]model.Counter, error)
	Set(name string, value model.Counter) error
	Delete(name string) error
}
