package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"reflect"
	"runtime"

	"github.com/7StaSH7/gometrics/internal/config"
	"github.com/7StaSH7/gometrics/internal/model"
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
	client  *http.Client
	baseURL string
}

type AgentInterface interface {
	GetMetric()
	SendMetrics()
	SendMetricsBatch()
	Close() error
}

func New(aCfg *config.AgentConfig) AgentInterface {
	return &Agent{
		client:  &http.Client{},
		baseURL: fmt.Sprintf("http://%s", aCfg.Address),
	}
}

func (a *Agent) SendMetrics() {
	v := reflect.ValueOf(&m)
	v = v.Elem()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		switch f.Kind() {
		case reflect.Float64:
			if err := a.sendOneMetric(model.Gauge, v.Type().Field(i).Name, f.Float()); err != nil {
				fmt.Printf("Error sending gauge metric %s: %v\n", v.Type().Field(i).Name, err)
			}

		case reflect.Int64:
			if err := a.sendOneMetric(model.Counter, v.Type().Field(i).Name, f.Int()); err != nil {
				fmt.Printf("Error sending counter metric %s: %v\n", v.Type().Field(i).Name, err)
			}
		}
	}
}

func (a *Agent) SendMetricsBatch() {
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
				MType: model.Gauge,
				Delta: &delta,
			})
		}
	}

	if len(metrics) > 0 {
		if err := a.sendBatchMetrics(metrics); err != nil {
			fmt.Printf("Error sending metrics %v\n", err)
		}
	}
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
	return nil
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

	url := fmt.Sprintf("%s/update/", a.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (a *Agent) sendBatchMetrics(metrics []model.Metrics) error {
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	url := fmt.Sprintf("%s/updates/", a.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
