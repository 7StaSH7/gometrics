package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"runtime"
	"sync"
	"time"

	"github.com/7StaSH7/gometrics/internal/config"
	"github.com/7StaSH7/gometrics/internal/logger"
	"github.com/7StaSH7/gometrics/internal/model"
	"github.com/7StaSH7/gometrics/internal/utils"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"resty.dev/v3"
)

type Gauge float64
type Counter int64

type MetricsMap map[string]any

type Agent struct {
	client  *resty.Client
	baseURL string

	ctx context.Context
	cfg *config.AgentConfig
	g   *errgroup.Group

	metrics   MetricsMap
	mu        sync.Mutex
	pollCount int64
	ms        runtime.MemStats
}

type AgentInterface interface {
	GetRuntimeMetrics() error
	GetGopsutilMetrics() error
	SendMetrics() error
	SendMetricsBatch() error
	Close() error
	Start(chan func() error)
}

func New(ctx context.Context, group *errgroup.Group, cfg *config.AgentConfig) AgentInterface {
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

	return &Agent{
		client:  client,
		baseURL: fmt.Sprintf("http://%s", cfg.Address),
		cfg:     cfg,

		metrics: make(MetricsMap),
		ctx:     ctx,
		g:       group,
	}
}

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
		if err := a.sendBatchMetrics(metricsBatch); err != nil {
			return fmt.Errorf("error sending metrics %+v", err)
		}
	}

	return nil
}

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

func (a *Agent) Close() error {
	return a.client.Close()
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

	url := fmt.Sprintf("%s/update/", a.baseURL)
	req := a.client.NewRequest().
		SetBody(body).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip")

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

	url := fmt.Sprintf("%s/updates/", a.baseURL)
	req := a.client.NewRequest().
		SetBody(metrics).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip")

	if a.cfg.Key != "" {
		hash := utils.GenerateSHA256(string(jsonData), a.cfg.Key)
		req.SetHeader("HashSHA256", hash)
	}

	if _, err := req.Post(url); err != nil {
		return err
	}

	return nil
}
