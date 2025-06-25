package grpcserver

import (
	"context"
	"crypto/rsa"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "google.golang.org/protobuf/proto"

	"github.com/nekr0z/muhame/internal/crypt"
	"github.com/nekr0z/muhame/internal/proto"
)

// DecryptInterceptor returns a grpc.UnaryServerInterceptor that decrypts the
// request. If the private key is nil, the interceptor does nothing.
func DecryptInterceptor(privateKey *rsa.PrivateKey) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if privateKey == nil {
			return handler(ctx, req)
		}

		var in []byte

		switch r := req.(type) {
		case *proto.MetricRequest:
			in = r.GetData()
		case *proto.BulkRequest:
			in = r.GetData()
		default:
			return nil, status.Error(codes.InvalidArgument, "invalid request type")
		}

		mm, err := crypt.Decrypt(in, privateKey)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		var m pb.Message

		switch r := req.(type) {
		case *proto.MetricRequest:
			m = r
		case *proto.BulkRequest:
			m = r
		}

		err = pb.Unmarshal(mm, m)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return handler(ctx, m)
	}
}
