package grpcserver_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	pb "google.golang.org/protobuf/proto"

	"github.com/nekr0z/muhame/internal/grpcserver"
	"github.com/nekr0z/muhame/internal/hash"
	"github.com/nekr0z/muhame/internal/metrics"
	"github.com/nekr0z/muhame/internal/storage"
	"github.com/nekr0z/muhame/pkg/proto"
)

func TestSignature(t *testing.T) {
	t.Parallel()

	key := "testkey"

	t.Run("update", func(t *testing.T) {
		t.Parallel()

		st, err := storage.New(zaptest.NewLogger(t).Sugar(), storage.Config{
			InMemory: true,
		})
		require.NoError(t, err)

		srv := grpcserver.New(st)
		cl := client(t, srv, grpc.UnaryInterceptor(grpcserver.SignatureInterceptor(key)))

		mr := &proto.MetricRequest{
			Payload: &proto.MetricRequest_Metric{
				Metric: &proto.Metric{
					Name: "test",
					Value: &proto.Metric_Counter{
						Counter: &proto.Counter{
							Delta: 1,
						},
					},
				},
			},
		}

		t.Run("unsigned", func(t *testing.T) {
			ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs())
			_, err := cl.Update(ctx, mr)
			assert.Error(t, err)

			r, err := st.Get(context.Background(), metrics.Counter.Type(0), "test")
			assert.Error(t, err, r)
		})

		t.Run("no meta", func(t *testing.T) {
			ctx := context.Background()
			_, err := cl.Update(ctx, mr)
			assert.Error(t, err)

			r, err := st.Get(context.Background(), metrics.Counter.Type(0), "test")
			assert.Error(t, err, r)
		})

		t.Run("bad signature", func(t *testing.T) {
			ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(hash.Header, "bad"))
			_, err := cl.Update(ctx, mr)
			assert.Error(t, err)

			r, err := st.Get(context.Background(), metrics.Counter.Type(0), "test")
			assert.Error(t, err, r)
		})

		t.Run("signed single", func(t *testing.T) {
			in, err := pb.Marshal(mr)
			require.NoError(t, err)

			sig := hash.Signature([]byte(in), key)
			meta := metadata.Pairs(hash.Header, sig)

			ctx := metadata.NewOutgoingContext(context.Background(), meta)
			_, err = cl.Update(ctx, mr)
			assert.NoError(t, err)

			r, err := st.Get(context.Background(), metrics.Counter.Type(0), "test")
			assert.NoError(t, err, r)

			assert.Equal(t, metrics.Counter(1), r)
		})

	})

	t.Run("bulk update", func(t *testing.T) {
		srv := grpcserver.New(mockBU{})
		cl := client(t, srv, grpc.UnaryInterceptor(grpcserver.SignatureInterceptor(key)))
		mr := &proto.BulkRequest{
			Payload: &proto.BulkRequest_Metrics{
				Metrics: &proto.Metrics{
					Metrics: []*proto.Metric{
						{
							Name: "test",
							Value: &proto.Metric_Counter{
								Counter: &proto.Counter{
									Delta: 1,
								},
							},
						},
						{
							Name: "test2",
							Value: &proto.Metric_Counter{
								Counter: &proto.Counter{
									Delta: 1,
								},
							},
						},
					},
				},
			},
		}

		in, err := pb.Marshal(mr)
		require.NoError(t, err)

		sig := hash.Signature([]byte(in), key)
		meta := metadata.Pairs(hash.Header, sig)

		ctx := metadata.NewOutgoingContext(context.Background(), meta)
		_, err = cl.BulkUpdate(ctx, mr)
		assert.NoError(t, err)
	})
}
