package postgres

import (
	"github.com/Dmitrevicz/gometrics/internal/model"
)

type CountersRepo struct {
	s *Storage
}

func NewCountersRepo(storage *Storage) *CountersRepo {
	return &CountersRepo{
		s: storage,
	}
}

func (s *CountersRepo) Get(name string) (model.Counter, bool) {
	// panic("not implemented")
	return 0, false // make autotests pass...
}

func (s *CountersRepo) GetAll() map[string]model.Counter {
	// panic("not implemented")
	return nil // make autotests pass...
}

func (s *CountersRepo) Set(name string, value model.Counter) {
	// panic("not implemented")
	// make autotests pass...
}

func (s *CountersRepo) Delete(name string) {
	// panic("not implemented")
	// make autotests pass...
}
