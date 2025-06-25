package grpcserver_test

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/nekr0z/muhame/internal/grpcserver"
	"github.com/nekr0z/muhame/internal/metrics"
	"github.com/nekr0z/muhame/internal/proto"
	"github.com/nekr0z/muhame/internal/storage"
)

func TestUpdate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := zaptest.NewLogger(t).Sugar()
	st, err := storage.New(log, storage.Config{InMemory: true})
	require.NoError(t, err)

	t.Run("happy path counter", func(t *testing.T) {
		t.Parallel()

		s := grpcserver.New(st)
		cl := client(t, s)

		m := &proto.Metric{
			Name: "testcounter",
			Value: &proto.Metric_Counter{
				Counter: &proto.Counter{
					Delta: 1,
				},
			},
		}

		_, err := cl.Update(ctx, &proto.MetricRequest{
			Payload: &proto.MetricRequest_Metric{
				Metric: m,
			},
		})
		assert.NoError(t, err)

		metric, err := st.Get(ctx, metrics.Counter.Type(0), "testcounter")
		assert.NoError(t, err)
		assert.Equal(t, metrics.Counter(1), metric)
	})

	t.Run("happy path gauge", func(t *testing.T) {
		t.Parallel()

		s := grpcserver.New(st)
		cl := client(t, s)

		m := &proto.Metric{
			Name: "testgauge",
			Value: &proto.Metric_Gauge{
				Gauge: &proto.Gauge{
					Value: 1.15,
				},
			},
		}

		_, err := cl.Update(ctx, &proto.MetricRequest{
			Payload: &proto.MetricRequest_Metric{
				Metric: m,
			},
		})
		assert.NoError(t, err)

		metric, err := st.Get(ctx, metrics.Gauge.Type(0), "testgauge")
		assert.NoError(t, err)
		assert.Equal(t, metrics.Gauge(1.15), metric)
	})

	t.Run("nil request", func(t *testing.T) {
		t.Parallel()

		s := grpcserver.New(st)
		cl := client(t, s)

		_, err := cl.Update(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("nil payload", func(t *testing.T) {
		t.Parallel()

		s := grpcserver.New(st)
		cl := client(t, s)

		_, err := cl.Update(ctx, &proto.MetricRequest{})
		assert.Error(t, err)
	})

	t.Run("nil metric", func(t *testing.T) {
		t.Parallel()

		s := grpcserver.New(st)
		cl := client(t, s)

		_, err := cl.Update(ctx, &proto.MetricRequest{
			Payload: &proto.MetricRequest_Metric{},
		})
		assert.Error(t, err)
	})

	t.Run("no name", func(t *testing.T) {
		t.Parallel()

		s := grpcserver.New(st)
		cl := client(t, s)

		_, err := cl.Update(ctx, &proto.MetricRequest{
			Payload: &proto.MetricRequest_Metric{
				Metric: &proto.Metric{},
			},
		})
		assert.Error(t, err)
	})

	t.Run("no metric", func(t *testing.T) {
		t.Parallel()

		s := grpcserver.New(st)
		cl := client(t, s)

		_, err := cl.Update(ctx, &proto.MetricRequest{
			Payload: &proto.MetricRequest_Metric{
				Metric: &proto.Metric{
					Name: "test",
				},
			}})
		assert.Error(t, err)
	})

	t.Run("failed update", func(t *testing.T) {
		t.Parallel()

		s := grpcserver.New(mockBU{})
		cl := client(t, s)

		m := &proto.Metric{
			Name: "fail",
			Value: &proto.Metric_Gauge{
				Gauge: &proto.Gauge{
					Value: 1.15,
				},
			},
		}

		_, err := cl.Update(ctx, &proto.MetricRequest{
			Payload: &proto.MetricRequest_Metric{
				Metric: m,
			},
		})
		assert.Error(t, err)
	})
}

func TestBulkUpdate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := zaptest.NewLogger(t).Sugar()

	t.Run("happy path", func(t *testing.T) {
		s := grpcserver.New(mockBU{})
		cl := client(t, s)

		_, err := cl.BulkUpdate(ctx, &proto.BulkRequest{
			Payload: &proto.BulkRequest_Metrics{
				Metrics: &proto.Metrics{
					Metrics: []*proto.Metric{
						{
							Name: "test",
							Value: &proto.Metric_Gauge{
								Gauge: &proto.Gauge{
									Value: 1.15,
								},
							},
						},
						{
							Name: "test",
							Value: &proto.Metric_Counter{
								Counter: &proto.Counter{
									Delta: 5,
								},
							},
						},
					},
				},
			},
		})
		assert.NoError(t, err)
	})

	t.Run("bad storage", func(t *testing.T) {
		st, err := storage.New(log, storage.Config{InMemory: true})
		require.NoError(t, err)

		s := grpcserver.New(st)
		cl := client(t, s)

		_, err = cl.BulkUpdate(ctx, &proto.BulkRequest{})
		assert.Error(t, err)
	})

	t.Run("nil request", func(t *testing.T) {
		s := grpcserver.New(mockBU{})
		cl := client(t, s)

		_, err := cl.BulkUpdate(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("no payload", func(t *testing.T) {
		s := grpcserver.New(mockBU{})
		cl := client(t, s)

		_, err := cl.BulkUpdate(ctx, &proto.BulkRequest{})
		assert.Error(t, err)
	})

	t.Run("no metrics", func(t *testing.T) {
		s := grpcserver.New(mockBU{})
		cl := client(t, s)

		_, err := cl.BulkUpdate(ctx, &proto.BulkRequest{
			Payload: &proto.BulkRequest_Metrics{
				Metrics: &proto.Metrics{},
			},
		})
		assert.Error(t, err)
	})

	t.Run("bad metrics", func(t *testing.T) {
		s := grpcserver.New(mockBU{})
		cl := client(t, s)

		_, err := cl.BulkUpdate(ctx, &proto.BulkRequest{
			Payload: &proto.BulkRequest_Metrics{
				Metrics: &proto.Metrics{
					Metrics: []*proto.Metric{
						{
							Name: "test",
						},
					},
				},
			},
		})
		assert.Error(t, err)
	})

	t.Run("failed update", func(t *testing.T) {
		s := grpcserver.New(mockBU{})
		cl := client(t, s)

		_, err := cl.BulkUpdate(ctx, &proto.BulkRequest{
			Payload: &proto.BulkRequest_Metrics{
				Metrics: &proto.Metrics{
					Metrics: []*proto.Metric{
						{
							Name: "test",
							Value: &proto.Metric_Gauge{
								Gauge: &proto.Gauge{
									Value: 1.15,
								},
							},
						},
					},
				},
			},
		})
		assert.Error(t, err)
	})
}

type mockBU struct{}

func (m mockBU) BulkUpdate(_ context.Context, mm []metrics.Named) error {
	if len(mm) == 1 {
		return assert.AnError
	}

	return nil
}

func (m mockBU) Update(_ context.Context, mm metrics.Named) error {
	if mm.Name == "fail" {
		return assert.AnError
	}

	return nil
}

func client(t *testing.T, ms *grpcserver.MetricsServer, opts ...grpc.ServerOption) proto.MetricsServiceClient {
	t.Helper()

	const bufSize = 1024 * 1024

	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer(opts...)

	proto.RegisterMetricsServiceServer(s, ms)

	go func() {
		err := s.Serve(lis)
		require.NoError(t, err)
	}()

	t.Cleanup(s.GracefulStop)

	conn, err := grpc.NewClient("passthrough:///bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	t.Cleanup(func() {
		err := conn.Close()
		assert.NoError(t, err)
	})

	return proto.NewMetricsServiceClient(conn)
}
