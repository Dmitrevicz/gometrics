package postgres

import (
	"database/sql"
	"errors"

	"github.com/Dmitrevicz/gometrics/internal/model"
	"github.com/Dmitrevicz/gometrics/internal/storage"
)

type CountersRepo struct {
	s *Storage
}

func NewCountersRepo(storage *Storage) *CountersRepo {
	return &CountersRepo{
		s: storage,
	}
}

const queryGetCounter = `SELECT value FROM counters WHERE name=$1;`

// Get finds metric by name. When requested metric doesn't exist
// storage.ErrNotFound error is returned.
func (r *CountersRepo) Get(name string) (model.Counter, error) {
	stmt, err := r.s.db.Prepare(queryGetCounter)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var counter model.Counter
	err = stmt.QueryRow(name).Scan(&counter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = storage.ErrNotFound
		}
		return 0, err
	}

	return counter, nil
}

const queryGetCountersAll = `SELECT name, value FROM counters;`

func (r *CountersRepo) GetAll() (map[string]model.Counter, error) {
	stmt, err := r.s.db.Prepare(queryGetCountersAll)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	counters := make(map[string]model.Counter)

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var (
			value model.Counter
			name  string
		)
		err = rows.Scan(&name, &value)
		if err != nil {
			return nil, err
		}
		counters[name] = value
	}

	return counters, rows.Err()
}

const querySetCounter = `
	INSERT INTO counters (name, value) VALUES ($1, $2)
	ON CONFLICT(name) 
	DO UPDATE SET value = counters.value + $2;
`

// Set updates the counter by its name or creates if doesn't exist.
//
// XXX: Инкремент #13. Использую `INSERT...ON CONFLICT DO UPDATE`, поэтому нет смысла
// проверять на pgerrcode.UniqueViolation.
func (r *CountersRepo) Set(name string, value model.Counter) error {
	stmt, err := r.s.db.Prepare(querySetCounter)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, value)

	return err
}

func (r *CountersRepo) BatchUpdate(counters []model.MetricCounter) (err error) {
	tx, err := r.s.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.Prepare(querySetCounter)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, counter := range counters {
		_, err = stmt.Exec(counter.Name, counter.Value)
		if err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

const queryDeleteCounter = `DELETE FROM counters WHERE name=$1;`

func (r *CountersRepo) Delete(name string) error {
	stmt, err := r.s.db.Prepare(queryDeleteCounter)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(name)

	return err
}
