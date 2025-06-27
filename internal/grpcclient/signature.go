// Package grpcclient provides client interceptors.
package grpcclient

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	pb "google.golang.org/protobuf/proto"

	"github.com/nekr0z/muhame/internal/hash"
	"github.com/nekr0z/muhame/pkg/proto"
)

// SignatureInterceptor returns a grpc.UnaryClientInterceptor that signs the
// request. If the key is empty, the interceptor does nothing.
func SignatureInterceptor(key string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if key == "" {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		var m pb.Message
		switch req := req.(type) {
		case *proto.MetricRequest:
			m = req
		case *proto.BulkRequest:
			m = req
		}

		out, err := pb.Marshal(m)
		if err != nil {
			return err
		}

		sig := hash.Signature(out, key)

		ctx = metadata.AppendToOutgoingContext(ctx, hash.Header, sig)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
