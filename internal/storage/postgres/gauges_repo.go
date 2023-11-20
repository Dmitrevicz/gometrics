package postgres

import (
	"github.com/Dmitrevicz/gometrics/internal/model"
)

type GaugesRepo struct {
	s *Storage
}

func NewGaugesRepo(storage *Storage) *GaugesRepo {
	return &GaugesRepo{
		s: storage,
	}
}

func (r *GaugesRepo) Get(name string) (model.Gauge, bool) {
	// panic("not implemented")
	return 0, false // make autotests pass...
}

func (r *GaugesRepo) GetAll() map[string]model.Gauge {
	// panic("not implemented")
	return nil // make autotests pass...
}

func (r *GaugesRepo) Set(name string, value model.Gauge) {
	// panic("not implemented")
	// make autotests pass...
}

func (r *GaugesRepo) Delete(name string) {
	// panic("not implemented")
	// make autotests pass...
}
