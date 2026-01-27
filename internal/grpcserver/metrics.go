// Package grpcserver provides gRPC handlers for metrics.
package grpcserver

import (
	"context"

	"github.com/7StaSH7/gometrics/internal/model"
	metricspb "github.com/7StaSH7/gometrics/internal/proto/metrics"
	"github.com/7StaSH7/gometrics/internal/service/metrics"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MetricsServer implements the gRPC Metrics service.
type MetricsServer struct {
	metricspb.UnimplementedMetricsServer
	metricsService metrics.MetricsService
}

// NewMetricsServer creates a new MetricsServer.
func NewMetricsServer(service metrics.MetricsService) *MetricsServer {
	return &MetricsServer{metricsService: service}
}

// UpdateMetrics updates metrics from gRPC requests.
func (s *MetricsServer) UpdateMetrics(ctx context.Context, req *metricspb.UpdateMetricsRequest) (*metricspb.UpdateMetricsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.GetMetrics()) == 0 {
		return &metricspb.UpdateMetricsResponse{}, nil
	}

	metricsBatch := make([]model.Metrics, 0, len(req.GetMetrics()))
	for _, metric := range req.GetMetrics() {
		if metric == nil {
			continue
		}
		if metric.GetId() == "" {
			return nil, status.Error(codes.InvalidArgument, "metric id is required")
		}

		switch metric.GetType() {
		case metricspb.Metric_GAUGE:
			value := metric.GetValue()
			metricsBatch = append(metricsBatch, model.Metrics{
				ID:    metric.GetId(),
				MType: model.Gauge,
				Value: &value,
			})
		case metricspb.Metric_COUNTER:
			delta := metric.GetDelta()
			metricsBatch = append(metricsBatch, model.Metrics{
				ID:    metric.GetId(),
				MType: model.Counter,
				Delta: &delta,
			})
		default:
			return nil, status.Error(codes.InvalidArgument, "unknown metric type")
		}
	}

	if len(metricsBatch) == 0 {
		return &metricspb.UpdateMetricsResponse{}, nil
	}

	if err := s.metricsService.Updates(ctx, metricsBatch); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &metricspb.UpdateMetricsResponse{}, nil
}
