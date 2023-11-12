package postgres

import (
	"database/sql"

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

var queryGetGauge = `SELECT value FROM gauges WHERE name=$1;`

// Get - bool result param shows if metric was found or not.
func (r *GaugesRepo) Get(name string) (model.Gauge, bool, error) {
	stmt, err := r.s.db.Prepare(queryGetGauge)
	if err != nil {
		return 0, false, err
	}
	defer stmt.Close()

	var gauge model.Gauge
	err = stmt.QueryRow(name).Scan(&gauge)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		return 0, false, err
	}

	return gauge, true, nil
}

var queryGetGaugesAll = `SELECT name, value FROM gauges;`

func (r *GaugesRepo) GetAll() (map[string]model.Gauge, error) {
	stmt, err := r.s.db.Prepare(queryGetGaugesAll)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	gauges := make(map[string]model.Gauge)

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var (
			value model.Gauge
			name  string
		)
		err = rows.Scan(&name, &value)
		if err != nil {
			return nil, err
		}
		gauges[name] = value
	}

	return gauges, rows.Err()
}

var querySetGauge = `
	INSERT INTO gauges (name, value) VALUES ($1, $2)
	ON CONFLICT(name) 
	DO UPDATE SET value=$2;
`

func (r *GaugesRepo) Set(name string, value model.Gauge) error {
	stmt, err := r.s.db.Prepare(querySetGauge)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, value)

	return err
}

var queryDeleteGauge = `DELETE FROM gauges WHERE name=$1;`

func (r *GaugesRepo) Delete(name string) error {
	stmt, err := r.s.db.Prepare(queryDeleteGauge)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(name)

	return err
}
