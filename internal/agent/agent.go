package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"reflect"
	"runtime"
	"time"

	"github.com/7StaSH7/gometrics/internal/config"
	"github.com/7StaSH7/gometrics/internal/model"
	"resty.dev/v3"
)

type Gauge float64
type Counter int64

type Metric struct {
	Alloc         Gauge
	BuckHashSys   Gauge
	Frees         Gauge
	GCCPUFraction Gauge
	GCSys         Gauge
	HeapAlloc     Gauge
	HeapIdle      Gauge
	HeapInuse     Gauge
	HeapObjects   Gauge
	HeapReleased  Gauge
	HeapSys       Gauge
	LastGC        Gauge
	Lookups       Gauge
	MCacheInuse   Gauge
	MCacheSys     Gauge
	MSpanInuse    Gauge
	MSpanSys      Gauge
	Mallocs       Gauge
	NextGC        Gauge
	NumForcedGC   Gauge
	NumGC         Gauge
	OtherSys      Gauge
	PauseTotalNs  Gauge
	StackInuse    Gauge
	StackSys      Gauge
	Sys           Gauge
	TotalAlloc    Gauge
	RandomValue   Gauge
	PollCount     Counter
}

var m Metric
var ms runtime.MemStats

type Agent struct {
	client  *resty.Client
	baseURL string
}

type AgentInterface interface {
	GetMetric()
	SendMetrics() error
	SendMetricsBatch() error
	Close() error
}

func New(cfg *config.AgentConfig) AgentInterface {
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
		SetDebug(true).
		SetRetryStrategy(
			func(_ *resty.Response, _ error) (time.Duration, error) {
				return 2 * time.Second, nil
			}).
		SetAllowNonIdempotentRetry(true).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second)

	return &Agent{
		client:  client,
		baseURL: fmt.Sprintf("http://%s", cfg.Address),
	}
}

func (a *Agent) SendMetrics() error {
	v := reflect.ValueOf(&m)
	v = v.Elem()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		switch f.Kind() {
		case reflect.Float64:
			if err := a.sendOneMetric(model.Gauge, v.Type().Field(i).Name, f.Float()); err != nil {
				return fmt.Errorf("Error sending gauge metric %s: %v", v.Type().Field(i).Name, err)
			}

		case reflect.Int64:
			if err := a.sendOneMetric(model.Counter, v.Type().Field(i).Name, f.Int()); err != nil {
				return fmt.Errorf("Error sending counter metric %s: %v", v.Type().Field(i).Name, err)
			}
		}
	}

	return nil
}

func (a *Agent) SendMetricsBatch() error {
	v := reflect.ValueOf(&m)
	v = v.Elem()
	metrics := make([]model.Metrics, 0, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		switch f.Kind() {
		case reflect.Float64:
			value := f.Float()
			metrics = append(metrics, model.Metrics{
				ID:    v.Type().Field(i).Name,
				MType: model.Gauge,
				Value: &value,
			})

		case reflect.Int64:
			delta := f.Int()
			metrics = append(metrics, model.Metrics{
				ID:    v.Type().Field(i).Name,
				MType: model.Counter,
				Delta: &delta,
			})
		}
	}

	if len(metrics) > 0 {
		if err := a.sendBatchMetrics(metrics); err != nil {
			return fmt.Errorf("Error sending metrics %v", err)
		}
	}

	return nil
}

func (a *Agent) GetMetric() {
	runtime.ReadMemStats(&ms)
	m.Alloc = Gauge(ms.Alloc)
	m.BuckHashSys = Gauge(ms.BuckHashSys)
	m.Frees = Gauge(ms.Frees)
	m.GCCPUFraction = Gauge(ms.GCCPUFraction)
	m.GCSys = Gauge(ms.GCSys)
	m.HeapAlloc = Gauge(ms.HeapAlloc)
	m.HeapIdle = Gauge(ms.HeapIdle)
	m.HeapInuse = Gauge(ms.HeapInuse)
	m.HeapObjects = Gauge(ms.HeapObjects)
	m.HeapReleased = Gauge(ms.HeapReleased)
	m.HeapSys = Gauge(ms.HeapSys)
	m.LastGC = Gauge(ms.LastGC)
	m.Lookups = Gauge(ms.Lookups)
	m.MCacheInuse = Gauge(ms.MCacheInuse)
	m.MCacheSys = Gauge(ms.MCacheSys)
	m.MSpanInuse = Gauge(ms.MSpanInuse)
	m.MSpanSys = Gauge(ms.MSpanSys)
	m.Mallocs = Gauge(ms.Mallocs)
	m.NextGC = Gauge(ms.NextGC)
	m.NumForcedGC = Gauge(ms.NumForcedGC)
	m.NumGC = Gauge(ms.NumGC)
	m.OtherSys = Gauge(ms.OtherSys)
	m.PauseTotalNs = Gauge(ms.PauseTotalNs)
	m.StackInuse = Gauge(ms.StackInuse)
	m.StackSys = Gauge(ms.StackSys)
	m.Sys = Gauge(ms.Sys)
	m.TotalAlloc = Gauge(ms.TotalAlloc)
	m.RandomValue = Gauge(rand.Float64())
	m.PollCount++
}

func (a *Agent) Close() error {
	return a.client.Close()
}

func (a *Agent) sendOneMetric(mType, name string, value any) error {
	body := model.Metrics{ID: name}
	switch mType {
	case model.Counter:
		{
			body.MType = model.Counter
			v, ok := value.(int64)
			if !ok {
				return errors.New("int64 not ok")
			}
			body.Delta = &v
		}
	case model.Gauge:
		{
			body.MType = model.Gauge
			v, ok := value.(float64)
			if !ok {
				return errors.New("float64 not ok")
			}
			body.Value = &v
		}
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Printf("send request with body: %s\n", string(jsonData))

	url := fmt.Sprintf("%s/update/", a.baseURL)
	req := a.client.NewRequest().
		SetBody(body).
		SetHeader("Content-Type", "application/json")

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
	fmt.Printf("send request with body: %s\n", string(jsonData))

	url := fmt.Sprintf("%s/updates/", a.baseURL)
	req := a.client.NewRequest().
		SetBody(metrics).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip")

	if _, err := req.Post(url); err != nil {
		return err
	}

	return nil
}
