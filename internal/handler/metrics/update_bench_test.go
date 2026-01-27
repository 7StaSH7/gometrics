package metrics

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/7StaSH7/gometrics/internal/audit"
	"github.com/7StaSH7/gometrics/internal/config"
	"github.com/7StaSH7/gometrics/internal/model"
	storagerepositsory "github.com/7StaSH7/gometrics/internal/repository/storage"
	metricsservice "github.com/7StaSH7/gometrics/internal/service/metrics"
	"github.com/7StaSH7/gometrics/internal/storage"
	"github.com/gin-gonic/gin"
)

func BenchmarkUpdateJSON(b *testing.B) {
	stor := storage.NewStorage(&config.ServerConfig{})
	storRep := storagerepositsory.NewMemStorageRepository(stor)
	mSer := metricsservice.New(storRep, nil)
	auditSubject := audit.NewAuditSubject()
	handler := New(mSer, "", "", auditSubject, "")

	router := gin.New()
	handler.Register(router)

	delta := int64(100)
	metric := model.Metrics{
		ID:    "test_counter",
		MType: model.Counter,
		Delta: &delta,
	}

	jsonData, _ := json.Marshal(metric)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkUpdateJSONWithHash(b *testing.B) {
	stor := storage.NewStorage(&config.ServerConfig{})
	storRep := storagerepositsory.NewMemStorageRepository(stor)
	mSer := metricsservice.New(storRep, nil)
	auditSubject := audit.NewAuditSubject()
	handler := New(mSer, "secret-key", "", auditSubject, "")

	router := gin.New()
	handler.Register(router)

	delta := int64(100)
	metric := model.Metrics{
		ID:    "test_counter",
		MType: model.Counter,
		Delta: &delta,
	}

	jsonData, _ := json.Marshal(metric)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("HashSHA256", "dummy-hash")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkUpdates(b *testing.B) {
	stor := storage.NewStorage(&config.ServerConfig{})
	storRep := storagerepositsory.NewMemStorageRepository(stor)
	mSer := metricsservice.New(storRep, nil)
	auditSubject := audit.NewAuditSubject()
	handler := New(mSer, "", "", auditSubject, "")

	router := gin.New()
	handler.Register(router)

	metrics := make([]model.Metrics, 0, 10)
	for i := range 10 {
		delta := int64(i * 10)
		metrics = append(metrics, model.Metrics{
			ID:    "counter_" + string(rune('0'+i)),
			MType: model.Counter,
			Delta: &delta,
		})
	}

	jsonData, _ := json.Marshal(metrics)

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkUpdateCounter(b *testing.B) {
	stor := storage.NewStorage(&config.ServerConfig{})
	storRep := storagerepositsory.NewMemStorageRepository(stor)
	mSer := metricsservice.New(storRep, nil)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_ = mSer.UpdateCounter(ctx, nil, "test_counter", 1)
	}
}

func BenchmarkUpdateGauge(b *testing.B) {
	stor := storage.NewStorage(&config.ServerConfig{})
	storRep := storagerepositsory.NewMemStorageRepository(stor)
	mSer := metricsservice.New(storRep, nil)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_ = mSer.UpdateGauge(ctx, nil, "test_gauge", 123.45)
	}
}
