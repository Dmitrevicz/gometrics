// Package implements grpc server methods for metrics.
package grpc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Dmitrevicz/gometrics/internal/logger"
	"github.com/Dmitrevicz/gometrics/internal/model"
	"github.com/Dmitrevicz/gometrics/internal/retry"
	"github.com/Dmitrevicz/gometrics/internal/server"
	"github.com/Dmitrevicz/gometrics/internal/server/config"
	pb "github.com/Dmitrevicz/gometrics/internal/server/grpc/proto"
	"github.com/Dmitrevicz/gometrics/internal/storage"
	"github.com/Dmitrevicz/gometrics/internal/storage/memstorage"
	"github.com/Dmitrevicz/gometrics/internal/storage/postgres"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// MetricsServer implements service methods.
type MetricsServer struct {
	pb.UnimplementedMetricsServer

	cfg     *config.Config
	Storage storage.Storage
}

func NewMetricsServer(cfg *config.Config) *MetricsServer {
	s := &MetricsServer{
		cfg: cfg,
	}

	// FIXME: pass storage as a dependency
	s.configureStorage(cfg)

	return s
}

// TODO: move storage setup away from server code
func (s *MetricsServer) configureStorage(cfg *config.Config) {
	if cfg.DatabaseDSN != "" {
		db, err := newDB(cfg.DatabaseDSN, true)
		if err != nil {
			logger.Log.Fatal("Can't configure storage", zap.Error(err))
		}

		// if err = createTables(db); err != nil {
		// 	logger.Log.Fatal("Can't configure storage", zap.Error(err))
		// }

		s.Storage = postgres.New(db)
		return
	}

	s.Storage = memstorage.New()
}

// TODO: move storage setup away from server code
func newDB(dsn string, withRetry bool) (db *sql.DB, err error) {
	var (
		retryInterval time.Duration
		retries       int
	)

	if withRetry {
		retryInterval = time.Second
		retries = 3
	}

	retry := retry.NewRetrier(retryInterval, retries)
	err = retry.Do("db open", func() error {
		db, err = sql.Open("pgx", dsn)
		if err != nil {
			if postgres.CheckRetriableErrors(err) {
				err = model.NewRetriableError(err)
			}
			return err
		}

		if err = db.Ping(); err != nil {
			if postgres.CheckRetriableErrors(err) {
				err = model.NewRetriableError(err)
			}
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return db, nil
}

func (s *MetricsServer) GetValue(ctx context.Context, req *pb.GetMetricRequest) (*pb.Metric, error) {
	var (
		metric pb.Metric
		value  interface{}
		err    error
	)

	// XXX: what should better be used - req.GetId() or req.Id?
	req.Id = strings.TrimSpace(req.Id)
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, server.ErrMsgEmptyMetricName)
	}

	switch strings.ToLower(req.Type.String()) {
	case model.MetricTypeGauge:
		value, err = s.Storage.Gauges().Get(req.Id)
		f := float64(value.(model.Gauge))
		metric.Value = &f
	case model.MetricTypeCounter:
		value, err = s.Storage.Counters().Get(req.Id)
		d := int64(value.(model.Counter))
		metric.Delta = &d
	default:
		return nil, status.Error(codes.InvalidArgument, server.ErrMsgWrongMetricType)
	}

	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Error(codes.NotFound, server.ErrMsgNothingFound)
		}

		logger.Log.Error(server.ErrMsgStorageFail, zap.Error(err))
		return nil, status.Error(codes.Internal, server.ErrMsgStorageFail)
	}

	metric.Id = req.Id
	metric.Type = req.Type

	return &metric, nil
}

func (s *MetricsServer) Update(ctx context.Context, req *pb.Metric) (*pb.Metric, error) {
	req.Id = strings.TrimSpace(req.Id)

	if req.Type == 0 {
		return nil, status.Error(codes.InvalidArgument, server.ErrMsgWrongMetricType)
	}

	if req.Id == "" {
		// http.StatusNotFound was required by specification in previous increments,
		// so return codes.NotFound now, I guess
		return nil, status.Error(codes.NotFound, server.ErrMsgEmptyMetricName)
	}

	var statusErr error
	switch strings.ToLower(req.Type.String()) {
	case model.MetricTypeGauge:
		statusErr = s.updateGauge(ctx, req)
	case model.MetricTypeCounter:
		statusErr = s.updateCounter(ctx, req)
	default:
		return nil, status.Error(codes.InvalidArgument, server.ErrMsgWrongMetricType)
	}

	if statusErr != nil {
		return nil, statusErr
	}

	return req, nil
}

func (s *MetricsServer) updateGauge(_ context.Context, m *pb.Metric) error {
	if m.Value == nil {
		// http.StatusNotFound was required in previous increments
		return status.Error(codes.NotFound, server.ErrMsgWrongMetricValue)
	}

	err := s.Storage.Gauges().Set(m.Id, model.Gauge(*m.Value))
	if err != nil {
		logger.Log.Error(server.ErrMsgStorageFail, zap.Error(err))
		return status.Error(codes.Internal, server.ErrMsgStorageFail)
	}
	m.Delta = nil

	// TODO: do smth with Dumper later

	return nil
}

func (s *MetricsServer) updateCounter(_ context.Context, m *pb.Metric) error {
	if m.Delta == nil {
		// http.StatusNotFound was required in previous increments
		return status.Error(codes.NotFound, server.ErrMsgWrongMetricValue)
	}

	if *m.Delta < 0 {
		return status.Error(codes.InvalidArgument, server.ErrMsgNegativeCounter)
	}

	err := s.Storage.Counters().Set(m.Id, model.Counter(*m.Delta))
	if err != nil {
		logger.Log.Error(server.ErrMsgStorageFail, zap.Error(err))
		return status.Error(codes.Internal, server.ErrMsgStorageFail)
	}

	// TODO: do smth with Dumper later

	counter, err := s.Storage.Counters().Get(m.Id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			logger.Log.Error(server.ErrMsgNothingFound + " after update attempt")
			return status.Error(codes.NotFound, server.ErrMsgNothingFound)
		}

		logger.Log.Error(server.ErrMsgStorageFail+" after update attempt", zap.Error(err))
		return status.Error(codes.Internal, server.ErrMsgStorageFail)
	}

	f := float64(counter)
	m.Value = &f

	return nil
}

func (s *MetricsServer) UpdateBatch(ctx context.Context, req *pb.UpdateBatchRequest) (*pb.UpdateBatchResponse, error) {
	gauges, counters, err := prepareBatchedMetrics(req.GetMetrics())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	logger.Log.Info("batch parsed", zap.Any("gauges", gauges), zap.Any("counters", counters))

	if err = s.Storage.Gauges().BatchUpdate(gauges); err != nil {
		logger.Log.Error(server.ErrMsgStorageFail, zap.Error(err))
		return nil, status.Error(codes.Internal, server.ErrMsgStorageFail)
	}

	if err = s.Storage.Counters().BatchUpdate(counters); err != nil {
		logger.Log.Error(server.ErrMsgStorageFail, zap.Error(err))
		return nil, status.Error(codes.Internal, server.ErrMsgStorageFail)
	}

	// TODO: do smth with Dumper later

	return &pb.UpdateBatchResponse{}, nil
}

// prepareBatchedMetrics prepares a batch of metrics for gauges and counters
// separately.
func prepareBatchedMetrics(metrics []*pb.Metric) (gs []model.MetricGauge, cs []model.MetricCounter, err error) {
	if len(metrics) == 0 {
		return
	}

	var mtype string
	for _, metric := range metrics {
		metric.Id = strings.TrimSpace(metric.Id)
		mtype = strings.ToLower(metric.Type.String())

		if metric.Type == 0 {
			return nil, nil, fmt.Errorf("%w: \"%s\"", server.ErrWrongMetricType, metric.Type)
		}
		if metric.Id == "" {
			return nil, nil, server.ErrEmptyMetricName
		}

		switch mtype {
		case model.MetricTypeGauge:
			if metric.Value == nil {
				return nil, nil, server.ErrWrongMetricValue
			}

			gs = append(gs, model.MetricGauge{
				Name:  metric.Id,
				Value: model.Gauge(*metric.Value),
			})
		case model.MetricTypeCounter:
			if metric.Delta == nil {
				return nil, nil, server.ErrWrongMetricValue
			}

			if *metric.Delta < 0 {
				return nil, nil, server.ErrNegativeCounter
			}

			cs = append(cs, model.MetricCounter{
				Name:  metric.Id,
				Value: model.Counter(*metric.Delta),
			})
		default:
			return nil, nil, fmt.Errorf("%w: \"%s\"", server.ErrWrongMetricType, metric.Type)
		}
	}

	return
}

func (s *MetricsServer) Ping(ctx context.Context, req *emptypb.Empty) (*pb.PingResponse, error) {
	if err := s.Storage.Ping(ctx); err != nil {
		logger.Log.Error("database ping failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "ping failure")
	}

	return &pb.PingResponse{
		Status: "ok",
	}, nil
}
