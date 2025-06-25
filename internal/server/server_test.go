package server

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/nekr0z/muhame/internal/addr"
	"github.com/nekr0z/muhame/internal/proto"
	"github.com/nekr0z/muhame/internal/storage"
)

func TestRun_http(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	t.Cleanup(wg.Wait)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)

	zapCore, observed := observer.New(zap.DebugLevel)

	cfg := config{
		address: addr.NetAddress{Host: "localhost", Port: 23456},
		st:      storage.Config{InMemory: true},
		log:     zap.New(zapCore),
	}

	go func() {
		err := run(ctx, cfg)
		require.NoError(t, err)
		wg.Done()
	}()

wait:
	for {
		select {
		case <-ctx.Done():
			t.FailNow()
		default:
			f := observed.Filter(func(e observer.LoggedEntry) bool {
				return strings.Contains(e.Message, "running server on")
			})
			if f.Len() > 0 {
				break wait
			}
			time.Sleep(100 * time.Millisecond)
		}
	}

	req, err := http.NewRequest("POST", "http://localhost:23456/update/gauge/test/1.2", nil)
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	defer func() {
		err = resp.Body.Close()
		assert.NoError(t, err)
	}()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req, err = http.NewRequest("GET", "http://localhost:23456/value/gauge/test", nil)
	assert.NoError(t, err)

	resp, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)

	defer func() {
		err = resp.Body.Close()
		assert.NoError(t, err)
	}()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	ans, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "1.2", string(ans))
}

func TestRun_grpc(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	t.Cleanup(wg.Wait)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)

	zapCore, observed := observer.New(zap.DebugLevel)

	cfg := config{
		address:     addr.NetAddress{Host: "localhost", Port: 23459},
		st:          storage.Config{InMemory: true},
		log:         zap.New(zapCore),
		gRPCaddress: addr.NetAddress{Host: "localhost", Port: 23458},
	}

	go func() {
		err := run(ctx, cfg)
		require.NoError(t, err)
		wg.Done()
	}()

wait:
	for {
		select {
		case <-ctx.Done():
			t.FailNow()
		default:
			f := observed.Filter(func(e observer.LoggedEntry) bool {
				return strings.Contains(e.Message, "running gRPC server on")
			})
			if f.Len() > 0 {
				break wait
			}
			time.Sleep(100 * time.Millisecond)
		}
	}

	conn, err := grpc.NewClient(":23458", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := proto.NewMetricsServiceClient(conn)

	res, err := client.Update(ctx, &proto.MetricRequest{
		Payload: &proto.MetricRequest_Metric{
			Metric: &proto.Metric{
				Name: "testcounter",
				Value: &proto.Metric_Counter{
					Counter: &proto.Counter{
						Delta: 1,
					},
				},
			},
		},
	})
	t.Log(res, err)
	assert.NoError(t, err)

	res, err = client.Update(ctx, &proto.MetricRequest{
		Payload: &proto.MetricRequest_Metric{
			Metric: &proto.Metric{
				Name: "testgauge",
				Value: &proto.Metric_Gauge{
					Gauge: &proto.Gauge{
						Value: 1.13,
					},
				},
			},
		},
	})
	t.Log(res, err)
	assert.NoError(t, err)

	mr, err := client.BulkUpdate(ctx, &proto.BulkRequest{
		Payload: &proto.BulkRequest_Metrics{
			Metrics: &proto.Metrics{
				Metrics: []*proto.Metric{
					{
						Name: "test",
						Value: &proto.Metric_Gauge{
							Gauge: &proto.Gauge{
								Value: 1.13,
							},
						},
					},
					{
						Name: "test",
						Value: &proto.Metric_Counter{
							Counter: &proto.Counter{
								Delta: 1,
							},
						},
					},
				},
			},
		},
	})
	t.Log(mr, err)
	assert.Error(t, err)
}
