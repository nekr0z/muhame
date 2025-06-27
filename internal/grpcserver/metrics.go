package grpcserver

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/nekr0z/muhame/internal/metrics"
	"github.com/nekr0z/muhame/pkg/proto"
)

// MetricsServer implements the grpc metrics service.
type MetricsServer struct {
	st Updater

	proto.UnimplementedMetricsServiceServer
}

// New returns a new MetricsServer.
func New(st Updater) *MetricsServer {
	return &MetricsServer{
		st: st,
	}
}

// Update implements the Update method.
func (s *MetricsServer) Update(ctx context.Context, in *proto.MetricRequest) (*emptypb.Empty, error) {
	if in == nil {
		return &emptypb.Empty{}, status.Error(codes.InvalidArgument, "no metric provided")
	}

	if in.GetMetric() == nil {
		return &emptypb.Empty{}, status.Error(codes.Internal, "not implemented")
	}

	m, msg := fromProto(in.GetMetric())
	if m == nil {
		return &emptypb.Empty{}, status.Error(codes.InvalidArgument, msg)
	}

	if err := s.st.Update(ctx, *m); err != nil {
		return &emptypb.Empty{}, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

// BulkUpdate implements the BulkUpdate method.
func (s *MetricsServer) BulkUpdate(ctx context.Context, in *proto.BulkRequest) (*emptypb.Empty, error) {
	bu, ok := s.st.(bulkUpdater)
	if !ok {
		return &emptypb.Empty{}, status.Error(codes.FailedPrecondition, "bulk update is not supported")
	}

	if in == nil {
		return &emptypb.Empty{}, status.Error(codes.InvalidArgument, "no metrics provided")
	}

	if in.GetMetrics() == nil {
		return &emptypb.Empty{}, status.Error(codes.Internal, "not implemented")
	}

	if len(in.GetMetrics().Metrics) == 0 {
		return &emptypb.Empty{}, status.Error(codes.InvalidArgument, "no metrics provided")
	}

	var ms []metrics.Named
	for _, m := range in.GetMetrics().Metrics {
		nm, msg := fromProto(m)
		if nm == nil {
			return &emptypb.Empty{}, status.Error(codes.InvalidArgument, strings.ToLower(msg))
		}

		ms = append(ms, *nm)
	}

	err := bu.BulkUpdate(ctx, ms)
	if err != nil {
		msg := err.Error()
		return &emptypb.Empty{}, status.Error(codes.Internal, msg)
	}

	return &emptypb.Empty{}, nil
}

func fromProto(in *proto.Metric) (*metrics.Named, string) {
	if in.GetName() == "" {
		return nil, "no metric name was provided"
	}

	m := metrics.Named{
		Name: in.GetName(),
	}

	switch in.GetValue().(type) {
	case *proto.Metric_Gauge:
		m.Metric = metrics.Gauge(in.GetGauge().GetValue())
	case *proto.Metric_Counter:
		m.Metric = metrics.Counter(in.GetCounter().GetDelta())
	default:
		return nil, "no metric of known type was provided"
	}

	return &m, ""
}

type Updater interface {
	Update(context.Context, metrics.Named) error
}

type bulkUpdater interface {
	BulkUpdate(context.Context, []metrics.Named) error
}
