package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Dmitrevicz/gometrics/internal/logger"
	"github.com/Dmitrevicz/gometrics/internal/model"
	"github.com/Dmitrevicz/gometrics/internal/server/config"
	"github.com/Dmitrevicz/gometrics/internal/storage"
	"go.uber.org/zap"
)

type Dumper struct {
	// quit    chan struct{}
	storage storage.Storage
	cfg     *config.Config
	mu      sync.Mutex

	// maybe add smth to be able to stop the ticker on Quit call
}

func NewDumper(storage storage.Storage, cfg *config.Config) *Dumper {
	return &Dumper{
		// quit:    make(chan struct{}),
		storage: storage,
		cfg:     cfg,
	}
}

// Start runs timer on specified interval.
// Can be stopped by call to Quit().
func (d *Dumper) Start() error {
	logger.Log.Info("Starting Dumper")

	// restore attempt
	if err := d.restore(); err != nil {
		return fmt.Errorf("unsuccessful restore attempt: %w", err)
	}

	// go d.waitForQuit()
	go d.startTimer()

	return nil
}

// func (d *Dumper) waitForQuit() {
// 	// go func() {
// 	<-d.quit
// 	logger.Log.Info("dumper caught quit")
// 	// }()
// }

func (d *Dumper) Quit() {
	logger.Log.Info("Stopping Dumper")

	if d.cfg.FileStoragePath == "" {
		logger.Log.Info("dump is disabled - empty file path")
		return
	}

	err := d.dump()
	if err != nil {
		logger.Log.Error("dumper got error trying to create a dump", zap.Error(err))
	}

	// close(d.quit)
}

func (d *Dumper) startTimer() {
	sleepDuration := time.Second * time.Duration(d.cfg.StoreInterval)

	// XXX: Не понял, что значит "значение 0 делает запись синхронной" в постановке задачи.
	// Поэтому просто отключаю в этом случае.
	if sleepDuration <= 0 {
		logger.Log.Error("dump timer wasn't started - got negative interval: " + sleepDuration.String())
		return
	}

	if d.cfg.FileStoragePath == "" {
		logger.Log.Info("dump timer is disabled - empty file path")
		return
	}

	logger.Log.Info("dumper timer started", zap.Duration(d.cfg.FileStoragePath, sleepDuration))

	for {
		time.Sleep(sleepDuration)

		err := d.dump()
		if err != nil {
			logger.Log.Error("dumper got error trying to create a dump", zap.Error(err))
		}
	}
}

type metricsDump struct {
	Gauges   map[string]model.Gauge   `json:"gauges"`
	Counters map[string]model.Counter `json:"counters"`
}

func (d *Dumper) dump() error {
	ts := time.Now()
	defer func() {
		logger.Log.Info("dumper dump took - " + time.Since(ts).String())
	}()

	logger.Log.Info("dump triggered")

	var metrics metricsDump

	metrics.Gauges = d.storage.Gauges().GetAll()
	metrics.Counters = d.storage.Counters().GetAll()

	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed json Marshal: %w", err)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	return os.WriteFile(d.cfg.FileStoragePath, data, 0666)
}

func (d *Dumper) restore() error {
	ts := time.Now()
	defer func() {
		logger.Log.Info("dumper restore took - " + time.Since(ts).String())
	}()

	logger.Log.Info("dump restore triggered")

	if !d.cfg.Restore {
		logger.Log.Info("dump restore is disabled by flag or env")
		return nil
	}

	if d.cfg.FileStoragePath == "" {
		logger.Log.Info("dump restore is disabled - empty file path")
		return nil
	}

	// read metrics from file
	d.mu.Lock()
	data, err := os.ReadFile(d.cfg.FileStoragePath)
	d.mu.Unlock()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Log.Warn("dumper didn't find a file to restore from (skipping restore)", zap.Error(err))
			return nil // what to do if file doesn't exist?
		}
		return fmt.Errorf("failed os.ReadFile: %w", err)
	}

	var metrics metricsDump
	if err := json.Unmarshal(data, &metrics); err != nil {
		return fmt.Errorf("failed json Unmarshal: %w", err)
	}

	// restore all metrics in storage
	counter := 0
	for name, value := range metrics.Counters {
		d.storage.Counters().Set(name, value)
		counter++
	}
	for name, value := range metrics.Gauges {
		d.storage.Gauges().Set(name, value)
		counter++
	}

	logger.Log.Info("restored metrics count: " + strconv.Itoa(counter))

	return nil
}
