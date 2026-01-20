package grpcserver

import (
	"context"
	"net"
	"sync"
	"testing"

	"github.com/7StaSH7/gometrics/internal/model"
	metricspb "github.com/7StaSH7/gometrics/internal/proto/metrics"
	"github.com/7StaSH7/gometrics/internal/service/metrics"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

type mockMetricsService struct {
	mu      sync.Mutex
	calls   int
	metrics []model.Metrics
}

func (m *mockMetricsService) UpdateCounter(ctx context.Context, tx pgx.Tx, name string, value int64) error {
	return nil
}

func (m *mockMetricsService) UpdateGauge(ctx context.Context, tx pgx.Tx, name string, value float64) error {
	return nil
}

func (m *mockMetricsService) GetCounter(name string) (int64, error) {
	return 0, nil
}

func (m *mockMetricsService) GetGauge(name string) (float64, error) {
	return 0, nil
}

func (m *mockMetricsService) GetMany() map[string]string {
	return nil
}

func (m *mockMetricsService) Store(ctx context.Context, restore bool, interval int) error {
	return nil
}

func (m *mockMetricsService) Updates(ctx context.Context, metrics []model.Metrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	m.metrics = append([]model.Metrics(nil), metrics...)
	return nil
}

func (m *mockMetricsService) snapshot() (int, []model.Metrics) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls, append([]model.Metrics(nil), m.metrics...)
}

func newTestClient(t *testing.T, interceptor grpc.UnaryServerInterceptor, service metrics.MetricsService) (metricspb.MetricsClient, func()) {
	t.Helper()

	listener := bufconn.Listen(bufSize)
	opts := []grpc.ServerOption{}
	if interceptor != nil {
		opts = append(opts, grpc.UnaryInterceptor(interceptor))
	}
	server := grpc.NewServer(opts...)
	metricspb.RegisterMetricsServer(server, NewMetricsServer(service))

	go func() {
		_ = server.Serve(listener)
	}()

	dialer := func(ctx context.Context, _ string) (net.Conn, error) {
		return listener.Dial()
	}

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		server.Stop()
		_ = listener.Close()
		t.Fatalf("failed to create gRPC client: %v", err)
	}

	cleanup := func() {
		_ = conn.Close()
		server.Stop()
		_ = listener.Close()
	}

	return metricspb.NewMetricsClient(conn), cleanup
}

func TestMetricsServer_UpdateMetrics(t *testing.T) {
	service := &mockMetricsService{}
	client, cleanup := newTestClient(t, nil, service)
	defer cleanup()

	metricsBatch := []*metricspb.Metric{
		metricspb.Metric_builder{Id: "Alloc", Type: metricspb.Metric_GAUGE, Value: 10.5}.Build(),
		metricspb.Metric_builder{Id: "PollCount", Type: metricspb.Metric_COUNTER, Delta: 7}.Build(),
	}
	req := metricspb.UpdateMetricsRequest_builder{Metrics: metricsBatch}.Build()

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("x-real-ip", "10.1.1.1"))
	if _, err := client.UpdateMetrics(ctx, req); err != nil {
		t.Fatalf("UpdateMetrics returned error: %v", err)
	}

	_, got := service.snapshot()
	if len(got) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(got))
	}

	if got[0].ID != "Alloc" || got[0].MType != model.Gauge || got[0].Value == nil || *got[0].Value != 10.5 {
		t.Fatalf("unexpected gauge metric: %#v", got[0])
	}
	if got[1].ID != "PollCount" || got[1].MType != model.Counter || got[1].Delta == nil || *got[1].Delta != 7 {
		t.Fatalf("unexpected counter metric: %#v", got[1])
	}
}

func TestTrustedSubnetInterceptor(t *testing.T) {
	service := &mockMetricsService{}
	interceptor := TrustedSubnetInterceptor("10.0.0.0/8")
	client, cleanup := newTestClient(t, interceptor, service)
	defer cleanup()

	req := metricspb.UpdateMetricsRequest_builder{
		Metrics: []*metricspb.Metric{
			metricspb.Metric_builder{Id: "Alloc", Type: metricspb.Metric_GAUGE, Value: 1.23}.Build(),
		},
	}.Build()

	allowedCtx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("x-real-ip", "10.1.2.3"))
	if _, err := client.UpdateMetrics(allowedCtx, req); err != nil {
		t.Fatalf("expected allowed request, got error: %v", err)
	}

	calls, _ := service.snapshot()
	if calls != 1 {
		t.Fatalf("expected 1 call after allowed request, got %d", calls)
	}

	deniedCtx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("x-real-ip", "192.168.1.10"))
	_, err := client.UpdateMetrics(deniedCtx, req)
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("expected permission denied, got %v", err)
	}

	calls, _ = service.snapshot()
	if calls != 1 {
		t.Fatalf("expected no new calls after denied request, got %d", calls)
	}
}
