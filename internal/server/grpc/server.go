// Package implements grpc server methods for metrics.
package grpc

import (
	"context"

	"github.com/Dmitrevicz/gometrics/internal/server/config"
	pb "github.com/Dmitrevicz/gometrics/internal/server/grpc/proto"
	"github.com/Dmitrevicz/gometrics/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// MetricsServer implements service methods.
type MetricsServer struct {
	pb.UnimplementedMetricsServer

	cfg     *config.Config
	storage storage.Storage
}

func NewMetricsServer(cfg *config.Config, storage storage.Storage) *MetricsServer {
	return &MetricsServer{
		cfg:     cfg,
		storage: storage,
	}
}

func (s *MetricsServer) GetValue(ctx context.Context, req *pb.GetMetricRequest) (*pb.GetMetricResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetValue not implemented")
}

func (s *MetricsServer) Update(ctx context.Context, req *pb.Metric) (*pb.Metric, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Update not implemented")
}

func (s *MetricsServer) UpdateBatch(ctx context.Context, req *pb.UpdateBatchRequest) (*pb.UpdateBatchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateBatch not implemented")
}

func (s *MetricsServer) Ping(ctx context.Context, req *emptypb.Empty) (*pb.PingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
