// Package agent provides the metrics agent that collects and sends metrics.
package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/7StaSH7/gometrics/internal/config"
	"github.com/7StaSH7/gometrics/internal/logger"
	"github.com/7StaSH7/gometrics/internal/model"
	metricspb "github.com/7StaSH7/gometrics/internal/proto/metrics"
	"github.com/7StaSH7/gometrics/internal/utils"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"resty.dev/v3"
)

// Gauge represents a gauge metric type.
type Gauge float64

// Counter represents a counter metric type.
type Counter int64

// MetricsMap is a map of metric names to their values.
type MetricsMap map[string]any

// Agent represents the metrics agent that collects and sends metrics.
type Agent struct {
	client  *resty.Client
	baseURL string

	grpcConn   *grpc.ClientConn
	grpcClient metricspb.MetricsClient
	localIP    string

	ctx context.Context
	cfg *config.AgentConfig
	g   *errgroup.Group

	metrics   MetricsMap
	mu        sync.Mutex
	pollCount int64
	ms        runtime.MemStats
}

// AgentInterface defines the interface for the metrics agent.
type AgentInterface interface {
	GetRuntimeMetrics() error
	GetGopsutilMetrics() error
	SendMetrics() error
	SendMetricsBatch() error
	Close() error
	Start(chan func() error)
}

// New creates a new Agent with the given context, errgroup, and config.
func New(ctx context.Context, group *errgroup.Group, cfg *config.AgentConfig) (AgentInterface, error) {
	client := resty.New().
		AddRetryConditions(
			func(res *resty.Response, err error) bool {
				if res == nil {
					return true
				}
				if err != nil {
					return true
				}
				return false
			},
		).
		SetContext(ctx).
		SetRetryStrategy(
			func(resp *resty.Response, _ error) (time.Duration, error) {
				select {
				case <-ctx.Done():
					return 0, ctx.Err()
				default:
				}
				var delay time.Duration
				switch resp.Request.Attempt {
				case 1:
					delay = 1 * time.Second
				case 2:
					delay = 3 * time.Second
				case 3:
					delay = 5 * time.Second
				default:
					delay = 5 * time.Second
				}
				logger.Log.Info("retrying", zap.Duration("delay", delay))
				return delay, nil
			}).
		SetAllowNonIdempotentRetry(true).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second)

	grpcConn, grpcClient, err := newGRPCClient(cfg.GRPCAddress)
	if err != nil {
		return nil, err
	}

	localIP := getLocalIP()

	return &Agent{
		client:     client,
		baseURL:    fmt.Sprintf("http://%s", cfg.Address),
		grpcConn:   grpcConn,
		grpcClient: grpcClient,
		localIP:    localIP,
		cfg:        cfg,

		metrics: make(MetricsMap),
		ctx:     ctx,
		g:       group,
	}, nil
}

// SendMetrics sends all collected metrics one by one.
func (a *Agent) SendMetrics() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for name, value := range a.metrics {
		switch v := value.(type) {
		case Gauge:
			if err := a.sendOneMetric(model.Gauge, name, float64(v)); err != nil {
				return fmt.Errorf("error sending gauge metric %s: %+v", name, err)
			}
		case Counter:
			if err := a.sendOneMetric(model.Counter, name, int64(v)); err != nil {
				return fmt.Errorf("error sending counter metric %s: %+v", name, err)
			}
		}
	}

	return nil
}

// SendMetricsBatch sends all collected metrics in batch.
func (a *Agent) SendMetricsBatch() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	metricsBatch := make([]model.Metrics, 0, len(a.metrics))
	for name, value := range a.metrics {
		switch v := value.(type) {
		case Gauge:
			val := float64(v)
			metricsBatch = append(metricsBatch, model.Metrics{
				ID:    name,
				MType: model.Gauge,
				Value: &val,
			})
		case Counter:
			delta := int64(v)
			metricsBatch = append(metricsBatch, model.Metrics{
				ID:    name,
				MType: model.Counter,
				Delta: &delta,
			})
		}
	}

	if len(metricsBatch) > 0 {
		if a.grpcClient != nil {
			if err := a.sendBatchMetricsGRPC(metricsBatch); err != nil {
				return fmt.Errorf("error sending gRPC metrics %+v", err)
			}
		} else {
			if err := a.sendBatchMetrics(metricsBatch); err != nil {
				return fmt.Errorf("error sending metrics %+v", err)
			}
		}
	}

	return nil
}

// GetRuntimeMetrics collects runtime metrics from Go.
func (a *Agent) GetRuntimeMetrics() error {
	runtime.ReadMemStats(&a.ms)
	a.mu.Lock()
	defer a.mu.Unlock()
	a.metrics["Alloc"] = Gauge(a.ms.Alloc)
	a.metrics["BuckHashSys"] = Gauge(a.ms.BuckHashSys)
	a.metrics["Frees"] = Gauge(a.ms.Frees)
	a.metrics["GCCPUFraction"] = Gauge(a.ms.GCCPUFraction)
	a.metrics["GCSys"] = Gauge(a.ms.GCSys)
	a.metrics["HeapAlloc"] = Gauge(a.ms.HeapAlloc)
	a.metrics["HeapIdle"] = Gauge(a.ms.HeapIdle)
	a.metrics["HeapInuse"] = Gauge(a.ms.HeapInuse)
	a.metrics["HeapObjects"] = Gauge(a.ms.HeapObjects)
	a.metrics["HeapReleased"] = Gauge(a.ms.HeapReleased)
	a.metrics["HeapSys"] = Gauge(a.ms.HeapSys)
	a.metrics["LastGC"] = Gauge(a.ms.LastGC)
	a.metrics["Lookups"] = Gauge(a.ms.Lookups)
	a.metrics["MCacheInuse"] = Gauge(a.ms.MCacheInuse)
	a.metrics["MCacheSys"] = Gauge(a.ms.MCacheSys)
	a.metrics["MSpanInuse"] = Gauge(a.ms.MSpanInuse)
	a.metrics["MSpanSys"] = Gauge(a.ms.MSpanSys)
	a.metrics["Mallocs"] = Gauge(a.ms.Mallocs)
	a.metrics["NextGC"] = Gauge(a.ms.NextGC)
	a.metrics["NumForcedGC"] = Gauge(a.ms.NumForcedGC)
	a.metrics["NumGC"] = Gauge(a.ms.NumGC)
	a.metrics["OtherSys"] = Gauge(a.ms.OtherSys)
	a.metrics["PauseTotalNs"] = Gauge(a.ms.PauseTotalNs)
	a.metrics["StackInuse"] = Gauge(a.ms.StackInuse)
	a.metrics["StackSys"] = Gauge(a.ms.StackSys)
	a.metrics["Sys"] = Gauge(a.ms.Sys)
	a.metrics["TotalAlloc"] = Gauge(a.ms.TotalAlloc)
	a.metrics["RandomValue"] = Gauge(rand.Float64())
	a.pollCount++
	a.metrics["PollCount"] = Counter(a.pollCount)

	return nil
}

// GetGopsutilMetrics collects system metrics using gopsutil.
func (a *Agent) GetGopsutilMetrics() error {
	v, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	c, err := cpu.Percent(0, true)
	if err != nil {
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	a.metrics["TotalMemory"] = Gauge(v.Total)
	a.metrics["FreeMemory"] = Gauge(v.Free)
	for i, cpuUtil := range c {
		a.metrics[fmt.Sprintf("CPUutilization%d", i)] = Gauge(cpuUtil)
	}

	return nil
}

// Start starts the agent with polling and reporting goroutines.
func (a *Agent) Start(sendJobs chan func() error) {
	a.g.Go(func() error {
		t := time.NewTicker(time.Duration(a.cfg.PollInterval) * time.Second)
		defer t.Stop()

		for {
			select {
			case <-a.ctx.Done():
				return a.ctx.Err()
			case <-t.C:
				if err := a.GetGopsutilMetrics(); err != nil {
					logger.Log.Error("get gopsutil metrics error", zap.Error(err))
				}
				if err := a.GetRuntimeMetrics(); err != nil {
					logger.Log.Error("get runtime metrics error", zap.Error(err))
				}
			}
		}
	})

	a.g.Go(func() error {
		ticker := time.NewTicker(time.Duration(a.cfg.ReportInterval) * time.Second)
		defer ticker.Stop()
		defer close(sendJobs)

		for {
			select {
			case <-a.ctx.Done():
				return a.ctx.Err()
			case <-ticker.C:
				sendJobs <- func() error {
					return a.SendMetricsBatch()
				}
			}
		}
	})
}

// Close closes the agent's gRPC and HTTP clients.
func (a *Agent) Close() error {
	var err error
	if a.grpcConn != nil {
		err = errors.Join(err, a.grpcConn.Close())
	}
	if a.client != nil {
		err = errors.Join(err, a.client.Close())
	}
	return err
}

func (a *Agent) sendOneMetric(mType, name string, value any) error {
	body := model.Metrics{ID: name}
	switch mType {
	case model.Counter:
		body.MType = model.Counter
		v, ok := value.(int64)
		if !ok {
			return errors.New("int64 not ok")
		}
		body.Delta = &v
	case model.Gauge:
		body.MType = model.Gauge
		v, ok := value.(float64)
		if !ok {
			return errors.New("float64 not ok")
		}
		body.Value = &v
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	logger.Log.Info("send request with body", zap.String("body", string(jsonData)))

	var requestBytes = jsonData
	if a.cfg.CryptoKey != "" {
		requestBytes, err = utils.Encrypt(jsonData)
		if err != nil {
			return fmt.Errorf("failed to encrypt data: %w", err)
		}
	}

	url := fmt.Sprintf("%s/update/", a.baseURL)
	req := a.client.NewRequest().
		SetBody(requestBytes).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("X-Real-IP", a.localIP)

	if a.cfg.Key != "" {
		hash := utils.GenerateSHA256(string(jsonData), a.cfg.Key)
		req.SetHeader("HashSHA256", hash)
	}

	if _, err := req.Post(url); err != nil {
		return err
	}

	return nil
}

func (a *Agent) sendBatchMetrics(metrics []model.Metrics) error {
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	logger.Log.Info("send request with body", zap.String("body", string(jsonData)))

	var requestBytes = jsonData
	if a.cfg.CryptoKey != "" {
		requestBytes, err = utils.Encrypt(jsonData)
		if err != nil {
			return fmt.Errorf("failed to encrypt data: %w", err)
		}
	}

	url := fmt.Sprintf("%s/updates/", a.baseURL)
	req := a.client.NewRequest().
		SetBody(requestBytes).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("X-Real-IP", a.localIP)

	if a.cfg.Key != "" {
		hash := utils.GenerateSHA256(string(jsonData), a.cfg.Key)
		req.SetHeader("HashSHA256", hash)
	}

	if _, err := req.Post(url); err != nil {
		return err
	}

	return nil
}

func (a *Agent) sendBatchMetricsGRPC(metrics []model.Metrics) error {
	if a.grpcClient == nil {
		return errors.New("grpc client is not configured")
	}

	pbMetrics := make([]*metricspb.Metric, 0, len(metrics))
	for _, metric := range metrics {
		pbMetric, err := toProtoMetric(metric)
		if err != nil {
			return err
		}
		pbMetrics = append(pbMetrics, pbMetric)
	}

	req := &metricspb.UpdateMetricsRequest{}
	req.SetMetrics(pbMetrics)
	ctx := metadata.NewOutgoingContext(a.ctx, metadata.Pairs("x-real-ip", a.localIP))
	if _, err := a.grpcClient.UpdateMetrics(ctx, req); err != nil {
		return err
	}

	return nil
}

func newGRPCClient(address string) (*grpc.ClientConn, metricspb.MetricsClient, error) {
	if address == "" {
		return nil, nil, nil
	}

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, errors.Join(fmt.Errorf("failed to connect to gRPC server %q: %w", address, err))
	}

	return conn, metricspb.NewMetricsClient(conn), nil
}

func toProtoMetric(metric model.Metrics) (*metricspb.Metric, error) {
	switch metric.MType {
	case model.Gauge:
		value := 0.0
		if metric.Value != nil {
			value = *metric.Value
		}
		pMetric := metricspb.Metric_builder{
			Id:    metric.ID,
			Type:  metricspb.Metric_GAUGE,
			Value: value,
		}.Build()

		return pMetric, nil
	case model.Counter:
		delta := int64(0)
		if metric.Delta != nil {
			delta = *metric.Delta
		}
		pMetric := metricspb.Metric_builder{
			Id:    metric.ID,
			Type:  metricspb.Metric_COUNTER,
			Value: float64(delta),
		}.Build()

		return pMetric, nil
	default:
		return nil, fmt.Errorf("unknown metric type %s", metric.MType)
	}
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}
