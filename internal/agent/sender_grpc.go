package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Dmitrevicz/gometrics/internal/agent/config"
	pb "github.com/Dmitrevicz/gometrics/internal/server/grpc/proto"
	"github.com/Dmitrevicz/gometrics/pkg/encryptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type grpcSender struct {
	sender        // inherit all fields from default sender for now...
	url    string // grpc server address
}

func NewSenderGRPC(cfg *config.Config, poller *poller, gopsutilPoller *gopsutilPoller) (*grpcSender, error) {
	log.Println("gRPC sender will be used")

	if cfg.RateLimit < 1 {
		cfg.RateLimit = 1
	}

	if cfg.HostIP == "" {
		log.Println("Couldn't detect host IP - X-Real-IP header will not be set on outgoing requests!")
	}

	var (
		encrypt *encryptor.Encryptor
		err     error
	)

	if cfg.CryptoKey != "" {
		encrypt, err = encryptor.NewEncryptor(cfg.CryptoKey)
		if err != nil {
			return nil, err
		}
	} else {
		log.Println("Empty CRYPTO_KEY was provided - encryption will be disabled!")
	}

	return &grpcSender{
		url: cfg.GRPCServerURL,
		sender: sender{
			reportInterval: cfg.ReportInterval,
			url:            cfg.ServerURL,
			key:            cfg.Key,
			hostIP:         cfg.HostIP,
			batch:          cfg.Batch,
			poller:         poller,
			gopsutilPoller: gopsutilPoller,
			client:         NewClientDefault(),
			Semaphore:      NewSemaphore(cfg.RateLimit),
			encryptor:      encrypt,
			quit:           make(chan struct{}),
		},
	}, nil
}

func (s *grpcSender) Start() {
	log.Println("Sender started")

	var ts time.Time
	sleepDuration := time.Second * time.Duration(s.reportInterval)
	s.timer = time.NewTimer(sleepDuration)

	for {
		select {
		case ts = <-s.timer.C:
			metrics := s.poller.AcquireMetrics()
			metrics.Merge(s.gopsutilPoller.AcquireMetrics())

			s.SendBatched(metrics)

			fmt.Println("send fired:", time.Since(ts))

			// I don't calculate delta-time, because current behaviour
			// is good enough right now.
			s.timer.Reset(sleepDuration)

		case <-s.quit:
			// stop the timer
			if !s.timer.Stop() {
				// drain the chanel (might not be needed here, but leave it be
				// as a kind of exercise)
				<-s.timer.C
			}

			log.Println("Sender timer stopped")
			return
		}
	}
}

// Shutdown stops sender's ticker and sends current data to server.
func (s *grpcSender) Shutdown(ctx context.Context) error {
	log.Println("Stopping Sender")

	s.stop()

	wait := make(chan error, 1)
	go func() {
		// send all data to the Server before program exit
		metrics := s.poller.AcquireMetrics()
		metrics.Merge(s.gopsutilPoller.AcquireMetrics())
		s.SendBatched(metrics)
		close(wait)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("sender Shutdown failed: %v", ctx.Err())
	case <-wait:
		return nil
	}
}

// stop stops sender's timer.
func (s *grpcSender) stop() {
	close(s.quit)
	log.Println("Sender stopped")
}

// DefaultGRPCClientTimeout - custom default timeout duration for gRPC client.
const DefaultGRPCClientTimeout = 10 * time.Second

// SendBatched overrides (overlaps) default sender behaviour to use gRPC client.
//
// TODO: compression, encryption, hash, host ip
func (s *grpcSender) SendBatched(metrics Metrics) {
	if metrics.Len() == 0 {
		log.Println("Metrics report skipped (nothing to be sent)")
		return
	}

	log.Println("Metrics report started (batched)")
	ts := time.Now()

	// XXX: Насколько плохо так делать? Как было бы правильней?
	// didn't have time to investigate how to reuse grpc connection properly, so
	// just create new every time, for now...
	conn, err := grpc.Dial(s.url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewMetricsClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultGRPCClientTimeout)
	defer cancel()

	req := new(pb.UpdateBatchRequest)
	s.prepareRequest(metrics, req)
	if _, err := client.UpdateBatch(ctx, req); err != nil {
		errMsg := "Got error while sending gRPC batched update request"
		e := status.Convert(err)
		log.Printf("%s: code: %s, err: %s\n", errMsg, e.Code(), e.Message())
	}

	log.Printf("Metrics (%d) have been sent (in %v)\n", len(req.Metrics), time.Since(ts))
}

func (s *grpcSender) prepareRequest(metrics Metrics, req *pb.UpdateBatchRequest) {
	if req == nil {
		log.Println("[grpcSender.prepareRequest] got nil as *pb.UpdateBatchRequest")
		return
	}

	req.Metrics = make([]*pb.Metric, 0, metrics.Len())

	for name, val := range metrics.Gauges {
		val := val
		gauge := pb.Metric{
			Type:  pb.MetricType_GAUGE,
			Id:    name,
			Value: (*float64)(&val),
		}
		req.Metrics = append(req.Metrics, &gauge)
	}

	for name, val := range metrics.Counters {
		val := val
		counter := pb.Metric{
			Type:  pb.MetricType_COUNTER,
			Id:    name,
			Delta: (*int64)(&val),
		}
		req.Metrics = append(req.Metrics, &counter)
	}
}
