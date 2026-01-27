package metrics_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"

	"github.com/7StaSH7/gometrics/internal/config"
	"github.com/7StaSH7/gometrics/internal/handler/health"
	metricsHandler "github.com/7StaSH7/gometrics/internal/handler/metrics"
	"github.com/7StaSH7/gometrics/internal/model"
	"github.com/7StaSH7/gometrics/internal/repository/storage"
	metricsService "github.com/7StaSH7/gometrics/internal/service/metrics"
	memStorage "github.com/7StaSH7/gometrics/internal/storage"
	"github.com/gin-gonic/gin"
)

// ExampleMetricsHandler_Update demonstrates updating a gauge metric via URL parameters.
func ExampleMetricsHandler_Update() {
	// Setup
	storageRepo := storage.NewMemStorageRepository(memStorage.NewStorage(&config.ServerConfig{}))
	service := metricsService.New(storageRepo, nil)
	handler := metricsHandler.New(service, "", "", nil, "")
	healthHandler := health.New(nil)

	r := gin.Default()
	handler.Register(r)
	healthHandler.Register(r)

	// Example request: update gauge metric
	req := httptest.NewRequest("POST", "/update/gauge/testGauge/123.45", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output: 200
}

// ExampleMetricsHandler_GetOne demonstrates retrieving a single metric via URL parameters.
func ExampleMetricsHandler_GetOne() {
	// Setup (similar to above)
	storageRepo := storage.NewMemStorageRepository(memStorage.NewStorage(&config.ServerConfig{}))
	service := metricsService.New(storageRepo, nil)
	handler := metricsHandler.New(service, "", "", nil, "")
	healthHandler := health.New(nil)

	r := gin.Default()
	handler.Register(r)
	healthHandler.Register(r)

	// First, update a metric
	req := httptest.NewRequest("POST", "/update/counter/testCounter/10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Then retrieve it
	req = httptest.NewRequest("GET", "/value/counter/testCounter", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fmt.Println(w.Code)
	fmt.Println(w.Body.String())
	// Output:
	// 200
	// 10
}

// ExampleMetricsHandler_UpdateJSON demonstrates updating a metric via JSON payload.
func ExampleMetricsHandler_UpdateJSON() {
	storageRepo := storage.NewMemStorageRepository(memStorage.NewStorage(&config.ServerConfig{}))
	service := metricsService.New(storageRepo, nil)
	handler := metricsHandler.New(service, "", "", nil, "")
	healthHandler := health.New(nil)

	r := gin.Default()
	handler.Register(r)
	healthHandler.Register(r)

	payload := model.Metrics{
		ID:    "testCounterJSON",
		MType: "counter",
		Delta: &[]int64{5}[0],
	}
	jsonData, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/update/", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output: 200
}

// ExampleMetricsHandler_GetJSON demonstrates retrieving a metric via JSON payload.
func ExampleMetricsHandler_GetJSON() {
	storageRepo := storage.NewMemStorageRepository(memStorage.NewStorage(&config.ServerConfig{}))
	service := metricsService.New(storageRepo, nil)
	handler := metricsHandler.New(service, "", "", nil, "")
	healthHandler := health.New(nil)

	r := gin.Default()
	handler.Register(r)
	healthHandler.Register(r)

	payload := model.Metrics{
		ID:    "testGaugeJSON",
		MType: "gauge",
		Value: &[]float64{67.89}[0],
	}
	jsonData, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/update/", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	getPayload := model.Metrics{
		ID:    "testGaugeJSON",
		MType: "gauge",
	}
	getData, _ := json.Marshal(getPayload)
	req = httptest.NewRequest("POST", "/value/", bytes.NewBuffer(getData))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output: 200
}

// ExampleMetricsHandler_Updates demonstrates updating multiple metrics in batch.
func ExampleMetricsHandler_Updates() {
	storageRepo := storage.NewMemStorageRepository(memStorage.NewStorage(&config.ServerConfig{}))
	service := metricsService.New(storageRepo, nil)
	handler := metricsHandler.New(service, "", "", nil, "")
	healthHandler := health.New(nil)

	r := gin.Default()
	handler.Register(r)
	healthHandler.Register(r)

	metricsBatch := []model.Metrics{
		{
			ID:    "batchCounter",
			MType: "counter",
			Delta: &[]int64{100}[0],
		},
		{
			ID:    "batchGauge",
			MType: "gauge",
			Value: &[]float64{99.99}[0],
		},
	}
	jsonData, _ := json.Marshal(metricsBatch)

	req := httptest.NewRequest("POST", "/updates/", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output: 200
}

// ExampleMetricsHandler_GetMany demonstrates retrieving all metrics in HTML format.
func ExampleMetricsHandler_GetMany() {
	storageRepo := storage.NewMemStorageRepository(memStorage.NewStorage(&config.ServerConfig{}))
	service := metricsService.New(storageRepo, nil)
	handler := metricsHandler.New(service, "", "", nil, "")
	healthHandler := health.New(nil)

	r := gin.Default()
	r.LoadHTMLGlob("../../../templates/*.tmpl")
	handler.Register(r)
	healthHandler.Register(r)

	req := httptest.NewRequest("POST", "/update/gauge/exampleGauge/42.0", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	req = httptest.NewRequest("POST", "/update/counter/exampleCounter/7", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	req = httptest.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fmt.Println(w.Code)
	// Output: 200
}
