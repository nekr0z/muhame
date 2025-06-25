package grpcserver

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	pb "google.golang.org/protobuf/proto"

	"github.com/nekr0z/muhame/internal/hash"
	"github.com/nekr0z/muhame/internal/proto"
)

// SignatureInterceptor returns a grpc.UnaryServerInterceptor that verifies the
// signature of the request. If the key is empty, the interceptor does nothing.
func SignatureInterceptor(key string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if key == "" {
			return handler(ctx, req)
		}

		var (
			in  []byte
			err error
		)

		switch r := req.(type) {
		case *proto.MetricRequest:
			in, err = pb.Marshal(r)
		case *proto.BulkRequest:
			in, err = pb.Marshal(r)
		default:
			return nil, status.Error(codes.InvalidArgument, "invalid request type")
		}

		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "failed to marshal request")
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "missing signature")
		}

		sig := md.Get(hash.Header)
		if len(sig) != 1 {
			return nil, status.Error(codes.InvalidArgument, "invalid signature")
		}

		if sig[0] != hash.Signature(in, key) {
			return nil, status.Error(codes.Unauthenticated, "invalid signature")
		}

		return handler(ctx, req)
	}
}
