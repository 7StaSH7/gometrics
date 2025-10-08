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
	"resty.dev/v3"
)

type Gauge float64
type Counter int64

type MetricsMap map[string]any

var metrics MetricsMap
var mu sync.Mutex
var pollCount int64
var ms runtime.MemStats

type Agent struct {
	client  *resty.Client
	baseURL string
	hashKey string
}

type AgentInterface interface {
	GetRuntimeMetrics() error
	GetGopsutilMetrics() error
	SendMetrics() error
	SendMetricsBatch() error
	Close() error
}

func New(ctx context.Context, cfg *config.AgentConfig) AgentInterface {
	metrics = make(MetricsMap)
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
		hashKey: cfg.Key,
	}
}

func (a *Agent) SendMetrics() error {
	mu.Lock()
	defer mu.Unlock()

	for name, value := range metrics {
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
	mu.Lock()
	defer mu.Unlock()

	metricsBatch := make([]model.Metrics, 0, len(metrics))
	for name, value := range metrics {
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
	runtime.ReadMemStats(&ms)
	mu.Lock()
	defer mu.Unlock()
	metrics["Alloc"] = Gauge(ms.Alloc)
	metrics["BuckHashSys"] = Gauge(ms.BuckHashSys)
	metrics["Frees"] = Gauge(ms.Frees)
	metrics["GCCPUFraction"] = Gauge(ms.GCCPUFraction)
	metrics["GCSys"] = Gauge(ms.GCSys)
	metrics["HeapAlloc"] = Gauge(ms.HeapAlloc)
	metrics["HeapIdle"] = Gauge(ms.HeapIdle)
	metrics["HeapInuse"] = Gauge(ms.HeapInuse)
	metrics["HeapObjects"] = Gauge(ms.HeapObjects)
	metrics["HeapReleased"] = Gauge(ms.HeapReleased)
	metrics["HeapSys"] = Gauge(ms.HeapSys)
	metrics["LastGC"] = Gauge(ms.LastGC)
	metrics["Lookups"] = Gauge(ms.Lookups)
	metrics["MCacheInuse"] = Gauge(ms.MCacheInuse)
	metrics["MCacheSys"] = Gauge(ms.MCacheSys)
	metrics["MSpanInuse"] = Gauge(ms.MSpanInuse)
	metrics["MSpanSys"] = Gauge(ms.MSpanSys)
	metrics["Mallocs"] = Gauge(ms.Mallocs)
	metrics["NextGC"] = Gauge(ms.NextGC)
	metrics["NumForcedGC"] = Gauge(ms.NumForcedGC)
	metrics["NumGC"] = Gauge(ms.NumGC)
	metrics["OtherSys"] = Gauge(ms.OtherSys)
	metrics["PauseTotalNs"] = Gauge(ms.PauseTotalNs)
	metrics["StackInuse"] = Gauge(ms.StackInuse)
	metrics["StackSys"] = Gauge(ms.StackSys)
	metrics["Sys"] = Gauge(ms.Sys)
	metrics["TotalAlloc"] = Gauge(ms.TotalAlloc)
	metrics["RandomValue"] = Gauge(rand.Float64())
	pollCount++
	metrics["PollCount"] = Counter(pollCount)

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

	mu.Lock()
	defer mu.Unlock()
	metrics["TotalMemory"] = Gauge(v.Total)
	metrics["FreeMemory"] = Gauge(v.Free)
	for i, cpuUtil := range c {
		metrics[fmt.Sprintf("CPUutilization%d", i)] = Gauge(cpuUtil)
	}

	return nil
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

	if a.hashKey != "" {
		hash := utils.GenerateSHA256(string(jsonData), a.hashKey)
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

	if a.hashKey != "" {
		hash := utils.GenerateSHA256(string(jsonData), a.hashKey)
		req.SetHeader("HashSHA256", hash)
	}

	if _, err := req.Post(url); err != nil {
		return err
	}

	return nil
}
