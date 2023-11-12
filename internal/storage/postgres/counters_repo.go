package postgres

import (
	"database/sql"

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

var queryGetCounter = `SELECT value FROM counters WHERE name=$1;`

// Get - bool result param shows if metric was found or not.
func (r *CountersRepo) Get(name string) (model.Counter, bool, error) {
	stmt, err := r.s.db.Prepare(queryGetCounter)
	if err != nil {
		return 0, false, err
	}
	defer stmt.Close()

	var counter model.Counter
	err = stmt.QueryRow(name).Scan(&counter)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		return 0, false, err
	}

	return counter, true, nil
}

var queryGetCountersAll = `SELECT name, value FROM counters;`

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

var querySetCounter = `
	INSERT INTO counters (name, value) VALUES ($1, $2)
	ON CONFLICT(name) 
	DO UPDATE SET value=$2;
`

func (r *CountersRepo) Set(name string, value model.Counter) error {
	stmt, err := r.s.db.Prepare(querySetCounter)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, value)

	return err
}

var queryDeleteCounter = `DELETE FROM counters WHERE name=$1;`

func (r *CountersRepo) Delete(name string) error {
	stmt, err := r.s.db.Prepare(queryDeleteCounter)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(name)

	return err
}
