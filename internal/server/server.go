package server

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/Dmitrevicz/gometrics/internal/logger"
	"github.com/Dmitrevicz/gometrics/internal/retry"
	"github.com/Dmitrevicz/gometrics/internal/server/config"
	"github.com/Dmitrevicz/gometrics/internal/storage"
	"github.com/Dmitrevicz/gometrics/internal/storage/memstorage"
	"github.com/Dmitrevicz/gometrics/internal/storage/postgres"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type server struct {
	router   http.Handler
	handlers *Handlers

	Storage storage.Storage
	Dumper  *Dumper
}

func New(cfg *config.Config) *server {
	var s server

	s.configureStorage(cfg)

	s.Dumper = NewDumper(s.Storage, cfg)
	s.handlers = NewHandlers(s.Storage, s.Dumper)

	// configure router
	gin.SetMode(gin.ReleaseMode)    // make it not spam logs on startup
	r := gin.New()                  // no middlewares
	r.RedirectTrailingSlash = false // autotests fail without it
	r.Use(gin.Recovery())           // only panic recover for now
	r.Use(RequestLogger())          // custom logger middleware from the lesson
	// r.Use(gin.Logger()) // gin.Logger can be used, but custom RequestLogger is preferred now in learning purposes

	r.Use(Gzip())

	// TODO: move routes configuration to separate func
	r.GET("/", s.handlers.PageIndex)
	r.GET("/all", s.handlers.GetAllMetrics)
	r.GET("/ping", s.handlers.PingStorage)
	r.GET("/value/:type/:name", s.handlers.GetMetricByName)
	r.POST("/value/", s.handlers.GetMetricByJSON)
	r.POST("/update/", s.handlers.UpdateMetricByJSON)
	r.POST("/update/:type/:name/:value", s.handlers.Update)
	r.POST("/updates/", s.handlers.UpdateBatch)
	// For endpoint "/update/:type/:name/:value" decided to use readable params
	// definition. Because instead you have to use *wildcard like "update/:type/*params"
	// or smth like this if needed to treat params errors more precisely

	s.router = r

	return &s
}

// requires router of type *http.ServeMux
// func (s *server) configureRouter() {
// 	s.Router.HandleFunc(`/update/`, internal.UpdateHandler)
// }

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) configureStorage(cfg *config.Config) {
	if cfg.DatabaseDSN != "" {
		db, err := newDB(cfg.DatabaseDSN, true)
		if err != nil {
			// или лучше прокидывать error вверх до самого main.go и уже там вызывать fatal?
			logger.Log.Fatal("Can't configure storage", zap.Error(err))
		}

		if err = createTables(db); err != nil {
			logger.Log.Fatal("Can't configure storage", zap.Error(err))
		}

		s.Storage = postgres.New(db)
		return
	}

	s.Storage = memstorage.New()
}

// XXX: куда можно положить эту функцию?
func newDB(dsn string, withRetry bool) (db *sql.DB, err error) {
	var (
		retryInterval time.Duration
		retries       int
	)

	if withRetry {
		retryInterval = time.Second
		retries = 3
	}

	// не совсем понял задание... попробовал навесить retry здесь...
	// но вроде это здесь не нужно
	retry := retry.NewRetrier(retryInterval, retries)
	err = retry.Do("db open", func() (error, bool) {
		db, err = sql.Open("pgx", dsn)
		if err != nil {
			return err, postgres.CheckRetriableErrors(err)
		}

		if err = db.Ping(); err != nil {
			return err, postgres.CheckRetriableErrors(err)
		}

		return nil, false
	})

	if err != nil {
		return nil, err
	}

	return db, nil
}

func createTables(db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err = tx.Exec(`CREATE TABLE IF NOT EXISTS counters (
		name varchar(500) NOT NULL PRIMARY KEY,
		value bigint NOT NULL DEFAULT 0,
		updated timestamptz NOT NULL DEFAULT now()
	)`); err != nil {
		return err
	}

	if _, err = tx.Exec(`CREATE TABLE IF NOT EXISTS gauges (
		name varchar(500) NOT NULL PRIMARY KEY,
		value double precision NOT NULL DEFAULT 0,
		updated timestamptz NOT NULL DEFAULT now()
	)`); err != nil {
		return err
	}

	return tx.Commit()
}
