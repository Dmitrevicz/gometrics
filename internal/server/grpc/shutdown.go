package grpc

import (
	"context"

	"google.golang.org/grpc"
)

func ShutdownWithContext(ctx context.Context, s *grpc.Server) error {
	wait := make(chan struct{}, 1)

	go func() {
		s.GracefulStop()
		close(wait)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-wait:
		return nil
	}
}
