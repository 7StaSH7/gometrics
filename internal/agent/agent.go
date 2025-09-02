package agent

import (
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"reflect"
	"runtime"

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
	*resty.Client
}

type AgentInterface interface {
	GetMetric()
	SendMetrics()
	Close() error
}

func New(aCfg *config.AgentConfig) AgentInterface {
	return &Agent{
		Client: resty.New().SetBaseURL(fmt.Sprintf("http://%s", aCfg.Address)),
	}
}

func (a *Agent) SendMetrics() {
	v := reflect.ValueOf(&m)
	v = v.Elem()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		switch f.Kind() {
		case reflect.Float64:
			if err := a.request(model.Gauge, v.Type().Field(i).Name, f.Float()); err != nil {
				panic(err)
			}
		case reflect.Int64:
			if err := a.request(model.Counter, v.Type().Field(i).Name, f.Int()); err != nil {
				panic(err)
			}
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
	return a.Client.Close()
}

func (a *Agent) request(mType, name string, value any) error {
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

	resp, err := a.Client.R().SetDebug(true).SetContentType("application/json").SetBody(body).SetCloseConnection(true).Post("/update/")
	if err != nil {
		fmt.Println("ERROR", err)
		return err
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return nil
}
