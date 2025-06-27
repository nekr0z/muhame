package grpcclient

import (
	"context"
	"crypto/rsa"

	"google.golang.org/grpc"
	pb "google.golang.org/protobuf/proto"

	"github.com/nekr0z/muhame/internal/crypt"
	"github.com/nekr0z/muhame/pkg/proto"
)

// EncryptInterceptor returns a grpc.UnaryClientInterceptor that encrypts the
// request. If the key is nil, the interceptor does nothing.
func EncryptInterceptor(key *rsa.PublicKey) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if key == nil {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		var (
			data []byte
			err  error
		)

		switch in := req.(type) {
		case *proto.MetricRequest:
			data, err = pb.Marshal(in)
		case *proto.BulkRequest:
			data, err = pb.Marshal(in)
		}

		if err != nil {
			return err
		}

		data, err = crypt.Encrypt(data, key)
		if err != nil {
			return err
		}

		switch req.(type) {
		case *proto.MetricRequest:
			req = &proto.MetricRequest{
				Payload: &proto.MetricRequest_Data{
					Data: data,
				},
			}
		case *proto.BulkRequest:
			req = &proto.BulkRequest{
				Payload: &proto.BulkRequest_Data{
					Data: data,
				},
			}
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
