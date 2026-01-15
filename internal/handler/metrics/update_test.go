package metrics

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"bytes"
	"os"

	"github.com/7StaSH7/gometrics/internal/audit"
	"github.com/7StaSH7/gometrics/internal/model"
	"github.com/7StaSH7/gometrics/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMetricsService struct {
	mock.Mock
}

type MockAuditObserver struct {
	NotifyCallCount int
	LastEvent       audit.AuditEvent
	Err             error
}

func (m *MockAuditObserver) Notify(ctx context.Context, event audit.AuditEvent) error {
	m.NotifyCallCount++
	m.LastEvent = event
	return m.Err
}

func (m *MockMetricsService) UpdateCounter(ctx context.Context, tx pgx.Tx, name string, value int64) error {
	args := m.Called(ctx, tx, name, value)

	return args.Error(0)
}

func (m *MockMetricsService) UpdateGauge(ctx context.Context, tx pgx.Tx, name string, value float64) error {
	args := m.Called(ctx, tx, name, value)

	return args.Error(0)
}

func (m *MockMetricsService) Updates(ctx context.Context, metrics []model.Metrics) error {
	args := m.Called(ctx, metrics)

	return args.Error(0)
}

func setupUpdateTestRouter(service *MockMetricsService, auditSubject *audit.AuditSubject) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := &metricsHandler{
		metricsService: service,
		audit:          auditSubject,
	}

	router.POST("/update/:type/:name/:value", handler.Update)

	return router
}

func setupUpdateJSONTestRouter(service *MockMetricsService, auditSubject *audit.AuditSubject) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := &metricsHandler{
		metricsService: service,
		audit:          auditSubject,
	}

	router.POST("/update/", handler.UpdateJSON)

	return router
}

func setupUpdatesTestRouter(service *MockMetricsService, auditSubject *audit.AuditSubject) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := &metricsHandler{
		metricsService: service,
		audit:          auditSubject,
	}

	router.POST("/updates/", handler.Updates)

	return router
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name                 string
		url                  string
		setupMock            func(*MockMetricsService)
		setupAudit           func(*MockAuditObserver)
		expectedStatus       int
		expectedHeader       string
		expectAuditCall      bool
		expectedAuditMetrics []string
	}{
		{
			name: "successful gauge update",
			url:  "/update/gauge/temperature/23.5",
			setupMock: func(m *MockMetricsService) {
				m.On("UpdateGauge", ctx, nil, "temperature", 23.5).Return(nil)
			},
			setupAudit: func(m *MockAuditObserver) {
				m.Err = nil
			},
			expectedStatus:       http.StatusOK,
			expectedHeader:       "text/plain; charset=utf-8",
			expectAuditCall:      true,
			expectedAuditMetrics: []string{"temperature"},
		},
		{
			name: "successful counter update",
			url:  "/update/counter/requests/100",
			setupMock: func(m *MockMetricsService) {
				m.On("UpdateCounter", ctx, nil, "requests", int64(100)).Return(nil)
			},
			setupAudit: func(m *MockAuditObserver) {
				m.Err = nil
			},
			expectedStatus:       http.StatusOK,
			expectedHeader:       "text/plain; charset=utf-8",
			expectAuditCall:      true,
			expectedAuditMetrics: []string{"requests"},
		},
		{
			name: "successful gauge update with negative value",
			url:  "/update/gauge/temperature/-15.3",
			setupMock: func(m *MockMetricsService) {
				m.On("UpdateGauge", ctx, nil, "temperature", -15.3).Return(nil)
			},
			setupAudit: func(m *MockAuditObserver) {
				m.Err = nil
			},
			expectedStatus:       http.StatusOK,
			expectedHeader:       "text/plain; charset=utf-8",
			expectAuditCall:      true,
			expectedAuditMetrics: []string{"temperature"},
		},
		{
			name: "successful gauge update with zero value",
			url:  "/update/gauge/pressure/0.0",
			setupMock: func(m *MockMetricsService) {
				m.On("UpdateGauge", ctx, nil, "pressure", 0.0).Return(nil)
			},
			setupAudit: func(m *MockAuditObserver) {
				m.Err = nil
			},
			expectedStatus:       http.StatusOK,
			expectedHeader:       "text/plain; charset=utf-8",
			expectAuditCall:      true,
			expectedAuditMetrics: []string{"pressure"},
		},
		{
			name: "successful counter update with zero value",
			url:  "/update/counter/errors/0",
			setupMock: func(m *MockMetricsService) {
				m.On("UpdateCounter", ctx, nil, "errors", int64(0)).Return(nil)
			},
			setupAudit: func(m *MockAuditObserver) {
				m.Err = nil
			},
			expectedStatus:       http.StatusOK,
			expectedHeader:       "text/plain; charset=utf-8",
			expectAuditCall:      true,
			expectedAuditMetrics: []string{"errors"},
		},
		{
			name:                 "invalid metric type",
			url:                  "/update/histogram/invalid/123",
			setupMock:            func(m *MockMetricsService) {},
			setupAudit:           func(m *MockAuditObserver) {},
			expectedStatus:       http.StatusBadRequest,
			expectedHeader:       "text/plain; charset=utf-8",
			expectAuditCall:      false,
			expectedAuditMetrics: nil,
		},
		{
			name:                 "empty metric name",
			url:                  "/update/gauge//123.45",
			setupMock:            func(m *MockMetricsService) {},
			setupAudit:           func(m *MockAuditObserver) {},
			expectedStatus:       http.StatusNotFound,
			expectedHeader:       "text/plain; charset=utf-8",
			expectAuditCall:      false,
			expectedAuditMetrics: nil,
		},
		{
			name:                 "invalid gauge value - not a float",
			url:                  "/update/gauge/temperature/invalid",
			setupMock:            func(m *MockMetricsService) {},
			setupAudit:           func(m *MockAuditObserver) {},
			expectedStatus:       http.StatusBadRequest,
			expectedHeader:       "text/plain; charset=utf-8",
			expectAuditCall:      false,
			expectedAuditMetrics: nil,
		},
		{
			name:                 "invalid counter value - not an integer",
			url:                  "/update/counter/requests/12.34",
			setupMock:            func(m *MockMetricsService) {},
			setupAudit:           func(m *MockAuditObserver) {},
			expectedStatus:       http.StatusBadRequest,
			expectedHeader:       "text/plain; charset=utf-8",
			expectAuditCall:      false,
			expectedAuditMetrics: nil,
		},
		{
			name:                 "invalid counter value - not a number",
			url:                  "/update/counter/requests/abc",
			setupMock:            func(m *MockMetricsService) {},
			setupAudit:           func(m *MockAuditObserver) {},
			expectedStatus:       http.StatusBadRequest,
			expectedHeader:       "text/plain; charset=utf-8",
			expectAuditCall:      false,
			expectedAuditMetrics: nil,
		},
		{
			name: "service error for gauge update",
			url:  "/update/gauge/temperature/25.0",
			setupMock: func(m *MockMetricsService) {
				m.On("UpdateGauge", ctx, nil, "temperature", 25.0).Return(errors.New("service error"))
			},
			setupAudit:           func(m *MockAuditObserver) {},
			expectedStatus:       http.StatusBadRequest,
			expectedHeader:       "text/plain; charset=utf-8",
			expectAuditCall:      false,
			expectedAuditMetrics: nil,
		},
		{
			name: "service error for counter update",
			url:  "/update/counter/requests/50",
			setupMock: func(m *MockMetricsService) {
				m.On("UpdateCounter", ctx, nil, "requests", int64(50)).Return(errors.New("service error"))
			},
			setupAudit:           func(m *MockAuditObserver) {},
			expectedStatus:       http.StatusBadRequest,
			expectedHeader:       "text/plain; charset=utf-8",
			expectAuditCall:      false,
			expectedAuditMetrics: nil,
		},
		{
			name: "large gauge value",
			url:  "/update/gauge/temperature/999999.999999",
			setupMock: func(m *MockMetricsService) {
				m.On("UpdateGauge", ctx, nil, "temperature", 999999.999999).Return(nil)
			},
			setupAudit: func(m *MockAuditObserver) {
				m.Err = nil
			},
			expectedStatus:       http.StatusOK,
			expectedHeader:       "text/plain; charset=utf-8",
			expectAuditCall:      true,
			expectedAuditMetrics: []string{"temperature"},
		},
		{
			name: "large counter value",
			url:  "/update/counter/requests/9223372036854775807",
			setupMock: func(m *MockMetricsService) {
				m.On("UpdateCounter", ctx, nil, "requests", int64(9223372036854775807)).Return(nil)
			},
			setupAudit: func(m *MockAuditObserver) {
				m.Err = nil
			},
			expectedStatus:       http.StatusOK,
			expectedHeader:       "text/plain; charset=utf-8",
			expectAuditCall:      true,
			expectedAuditMetrics: []string{"requests"},
		},
		{
			name:                 "counter value overflow",
			url:                  "/update/counter/requests/99999999999999999999",
			setupMock:            func(m *MockMetricsService) {},
			setupAudit:           func(m *MockAuditObserver) {},
			expectedStatus:       http.StatusBadRequest,
			expectedHeader:       "text/plain; charset=utf-8",
			expectAuditCall:      false,
			expectedAuditMetrics: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockMetricsService)
			tt.setupMock(mockService)

			mockAuditObserver := new(MockAuditObserver)
			tt.setupAudit(mockAuditObserver)

			auditSubject := audit.NewAuditSubject()
			auditSubject.Attach(mockAuditObserver)

			router := setupUpdateTestRouter(mockService, auditSubject)

			req, err := http.NewRequest(http.MethodPost, tt.url, nil)
			assert.NoError(t, err)
			// Устанавливаем IP адрес для тестирования
			req.RemoteAddr = "127.0.0.1:8080"

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Status code mismatch")
			assert.Equal(t, tt.expectedHeader, w.Header().Get("Content-Type"), "Content-Type header mismatch")

			mockService.AssertExpectations(t)

			if tt.expectAuditCall {
				assert.Equal(t, 1, mockAuditObserver.NotifyCallCount, "Audit should be called once")
				assert.Equal(t, tt.expectedAuditMetrics, mockAuditObserver.LastEvent.Metrics, "Audit metrics mismatch")
				assert.Equal(t, "127.0.0.1", mockAuditObserver.LastEvent.IPAddress, "Audit IP address mismatch")
				assert.Greater(t, mockAuditObserver.LastEvent.Timestamp, int64(0), "Audit timestamp should be set")
			} else {
				assert.Equal(t, 0, mockAuditObserver.NotifyCallCount, "Audit should not be called")
			}
		})
	}
}

func TestUpdateJSON(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name                 string
		body                 string
		setupMock            func(*MockMetricsService)
		setupAudit           func(*MockAuditObserver)
		expectedStatus       int
		expectAuditCall      bool
		expectedAuditMetrics []string
	}{
		{
			name: "successful gauge update",
			body: `{"id":"temperature","type":"gauge","value":23.5}`,
			setupMock: func(m *MockMetricsService) {
				m.On("UpdateGauge", ctx, nil, "temperature", 23.5).Return(nil)
			},
			setupAudit: func(m *MockAuditObserver) {
				m.Err = nil
			},
			expectedStatus:       http.StatusOK,
			expectAuditCall:      true,
			expectedAuditMetrics: []string{"temperature"},
		},
		{
			name: "successful counter update",
			body: `{"id":"requests","type":"counter","delta":100}`,
			setupMock: func(m *MockMetricsService) {
				m.On("UpdateCounter", ctx, nil, "requests", int64(100)).Return(nil)
			},
			setupAudit: func(m *MockAuditObserver) {
				m.Err = nil
			},
			expectedStatus:       http.StatusOK,
			expectAuditCall:      true,
			expectedAuditMetrics: []string{"requests"},
		},
		{
			name: "invalid JSON",
			body: `{"invalid": json}`,
			setupMock: func(m *MockMetricsService) {
			},
			setupAudit: func(m *MockAuditObserver) {
			},
			expectedStatus:       http.StatusBadRequest,
			expectAuditCall:      false,
			expectedAuditMetrics: nil,
		},
		{
			name: "service error",
			body: `{"id":"temperature","type":"gauge","value":23.5}`,
			setupMock: func(m *MockMetricsService) {
				m.On("UpdateGauge", ctx, nil, "temperature", 23.5).Return(errors.New("service error"))
			},
			setupAudit: func(m *MockAuditObserver) {
			},
			expectedStatus:       http.StatusBadRequest,
			expectAuditCall:      false,
			expectedAuditMetrics: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockMetricsService)
			tt.setupMock(mockService)

			mockAuditObserver := new(MockAuditObserver)
			tt.setupAudit(mockAuditObserver)

			auditSubject := audit.NewAuditSubject()
			auditSubject.Attach(mockAuditObserver)

			router := setupUpdateJSONTestRouter(mockService, auditSubject)

			req, err := http.NewRequest(http.MethodPost, "/update/", bytes.NewBufferString(tt.body))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.RemoteAddr = "127.0.0.1:8080"

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Status code mismatch")

			mockService.AssertExpectations(t)

			if tt.expectAuditCall {
				assert.Equal(t, 1, mockAuditObserver.NotifyCallCount, "Audit should be called once")
				assert.Equal(t, tt.expectedAuditMetrics, mockAuditObserver.LastEvent.Metrics, "Audit metrics mismatch")
				assert.Equal(t, "127.0.0.1", mockAuditObserver.LastEvent.IPAddress, "Audit IP address mismatch")
				assert.Greater(t, mockAuditObserver.LastEvent.Timestamp, int64(0), "Audit timestamp should be set")
			} else {
				assert.Equal(t, 0, mockAuditObserver.NotifyCallCount, "Audit should not be called")
			}
		})
	}
}

func TestUpdates(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name                 string
		body                 string
		setupMock            func(*MockMetricsService)
		setupAudit           func(*MockAuditObserver)
		expectedStatus       int
		expectAuditCall      bool
		expectedAuditMetrics []string
	}{
		{
			name: "successful batch update",
			body: `[{"id":"temperature","type":"gauge","value":23.5},{"id":"requests","type":"counter","delta":100}]`,
			setupMock: func(m *MockMetricsService) {
				metrics := []model.Metrics{
					{ID: "temperature", MType: "gauge", Value: float64Ptr(23.5)},
					{ID: "requests", MType: "counter", Delta: int64Ptr(100)},
				}
				m.On("Updates", ctx, metrics).Return(nil)
			},
			setupAudit: func(m *MockAuditObserver) {
				m.Err = nil
			},
			expectedStatus:       http.StatusOK,
			expectAuditCall:      true,
			expectedAuditMetrics: []string{"temperature", "requests"},
		},
		{
			name: "invalid JSON",
			body: `[{"invalid": json}]`,
			setupMock: func(m *MockMetricsService) {
			},
			setupAudit: func(m *MockAuditObserver) {
			},
			expectedStatus:       http.StatusBadRequest,
			expectAuditCall:      false,
			expectedAuditMetrics: nil,
		},
		{
			name: "service error",
			body: `[{"id":"temperature","type":"gauge","value":23.5}]`,
			setupMock: func(m *MockMetricsService) {
				metrics := []model.Metrics{
					{ID: "temperature", MType: "gauge", Value: float64Ptr(23.5)},
				}
				m.On("Updates", ctx, metrics).Return(errors.New("service error"))
			},
			setupAudit: func(m *MockAuditObserver) {
			},
			expectedStatus:       http.StatusInternalServerError,
			expectAuditCall:      false,
			expectedAuditMetrics: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockMetricsService)
			tt.setupMock(mockService)

			mockAuditObserver := new(MockAuditObserver)
			tt.setupAudit(mockAuditObserver)

			auditSubject := audit.NewAuditSubject()
			auditSubject.Attach(mockAuditObserver)

			router := setupUpdatesTestRouter(mockService, auditSubject)

			req, err := http.NewRequest(http.MethodPost, "/updates/", bytes.NewBufferString(tt.body))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.RemoteAddr = "127.0.0.1:8080"

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Status code mismatch")

			mockService.AssertExpectations(t)

			if tt.expectAuditCall {
				assert.Equal(t, 1, mockAuditObserver.NotifyCallCount, "Audit should be called once")
				assert.Equal(t, tt.expectedAuditMetrics, mockAuditObserver.LastEvent.Metrics, "Audit metrics mismatch")
				assert.Equal(t, "127.0.0.1", mockAuditObserver.LastEvent.IPAddress, "Audit IP address mismatch")
				assert.Greater(t, mockAuditObserver.LastEvent.Timestamp, int64(0), "Audit timestamp should be set")
			} else {
				assert.Equal(t, 0, mockAuditObserver.NotifyCallCount, "Audit should not be called")
			}
		})
	}
}

func float64Ptr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}

func setupEncryptedUpdateJSONTestRouter(service *MockMetricsService, auditSubject *audit.AuditSubject, hashKey, cryptoKey string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := &metricsHandler{
		metricsService: service,
		hashKey:        hashKey,
		cryptoKey:      cryptoKey,
		audit:          auditSubject,
	}

	router.POST("/update/", handler.UpdateJSON)

	return router
}

func setupEncryptedUpdatesTestRouter(service *MockMetricsService, auditSubject *audit.AuditSubject, hashKey, cryptoKey string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := &metricsHandler{
		metricsService: service,
		hashKey:        hashKey,
		cryptoKey:      cryptoKey,
		audit:          auditSubject,
	}

	router.POST("/updates/", handler.Updates)

	return router
}

func generateTestKeys() (*rsa.PrivateKey, string, string) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		panic(err)
	}
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	privateKeyFile := "/tmp/test_private_key_update_test.pem"
	if err := os.WriteFile(privateKeyFile, privateKeyPEM, 0600); err != nil {
		panic(err)
	}

	if _, err := utils.LoadPrivateKey(privateKeyFile); err != nil {
		panic(err)
	}

	return privateKey, privateKeyFile, ""
}

func encryptData(privateKey *rsa.PrivateKey, data []byte) []byte {
	hash := sha256.New()
	encrypted, err := rsa.EncryptOAEP(hash, rand.Reader, &privateKey.PublicKey, data, nil)
	if err != nil {
		panic(err)
	}
	return encrypted
}

func TestEncryptionInUpdateJSON(t *testing.T) {
	ctx := context.Background()
	privateKey, cryptoKeyPath, _ := generateTestKeys()

	tests := []struct {
		name                 string
		body                 []byte
		headers              map[string]string
		setupMock            func(*MockMetricsService)
		setupAudit           func(*MockAuditObserver)
		expectedStatus       int
		expectAuditCall      bool
		expectedAuditMetrics []string
	}{
		{
			name: "successful encrypted counter update",
			body: encryptData(privateKey, []byte(`{"id":"test_counter_encrypted","type":"counter","delta":100}`)),
			setupMock: func(m *MockMetricsService) {
				m.On("UpdateCounter", ctx, nil, "test_counter_encrypted", int64(100)).Return(nil)
			},
			setupAudit: func(m *MockAuditObserver) {
				m.Err = nil
			},
			expectedStatus:       http.StatusOK,
			expectAuditCall:      true,
			expectedAuditMetrics: []string{"test_counter_encrypted"},
		},
		{
			name: "successful encrypted gauge update",
			body: encryptData(privateKey, []byte(`{"id":"test_gauge_encrypted","type":"gauge","value":123.45}`)),
			setupMock: func(m *MockMetricsService) {
				m.On("UpdateGauge", ctx, nil, "test_gauge_encrypted", 123.45).Return(nil)
			},
			setupAudit: func(m *MockAuditObserver) {
				m.Err = nil
			},
			expectedStatus:       http.StatusOK,
			expectAuditCall:      true,
			expectedAuditMetrics: []string{"test_gauge_encrypted"},
		},
		{
			name:                 "invalid encrypted data",
			body:                 []byte("invalid encrypted data"),
			setupMock:            func(m *MockMetricsService) {},
			setupAudit:           func(m *MockAuditObserver) {},
			expectedStatus:       http.StatusBadRequest,
			expectAuditCall:      false,
			expectedAuditMetrics: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockMetricsService)
			tt.setupMock(mockService)

			mockAuditObserver := new(MockAuditObserver)
			tt.setupAudit(mockAuditObserver)

			auditSubject := audit.NewAuditSubject()
			auditSubject.Attach(mockAuditObserver)

			router := setupEncryptedUpdateJSONTestRouter(mockService, auditSubject, "test-hash-key", cryptoKeyPath)

			req, err := http.NewRequest(http.MethodPost, "/update/", bytes.NewReader(tt.body))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.RemoteAddr = "127.0.0.1:8080"

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Status code mismatch")

			mockService.AssertExpectations(t)

			if tt.expectAuditCall {
				assert.Equal(t, 1, mockAuditObserver.NotifyCallCount, "Audit should be called once")
				assert.Equal(t, tt.expectedAuditMetrics, mockAuditObserver.LastEvent.Metrics, "Audit metrics mismatch")
				assert.Equal(t, "127.0.0.1", mockAuditObserver.LastEvent.IPAddress, "Audit IP address mismatch")
				assert.Greater(t, mockAuditObserver.LastEvent.Timestamp, int64(0), "Audit timestamp should be set")
			} else {
				assert.Equal(t, 0, mockAuditObserver.NotifyCallCount, "Audit should not be called")
			}
		})
	}
}

func TestEncryptionInUpdates(t *testing.T) {
	ctx := context.Background()
	privateKey, cryptoKeyPath, _ := generateTestKeys()

	tests := []struct {
		name                 string
		body                 []byte
		setupMock            func(*MockMetricsService)
		setupAudit           func(*MockAuditObserver)
		expectedStatus       int
		expectAuditCall      bool
		expectedAuditMetrics []string
	}{
		{
			name: "successful encrypted batch update",
			body: encryptData(privateKey, []byte(`[{"id":"counter1","type":"counter","delta":50},{"id":"gauge1","type":"gauge","value":123.45}]`)),
			setupMock: func(m *MockMetricsService) {
				m.On("Updates", ctx, mock.Anything).Return(nil)
			},
			setupAudit: func(m *MockAuditObserver) {
				m.Err = nil
			},
			expectedStatus:       http.StatusOK,
			expectAuditCall:      true,
			expectedAuditMetrics: []string{"counter1", "gauge1"},
		},
		{
			name:                 "invalid encrypted batch data",
			body:                 []byte("invalid encrypted batch"),
			setupMock:            func(m *MockMetricsService) {},
			setupAudit:           func(m *MockAuditObserver) {},
			expectedStatus:       http.StatusBadRequest,
			expectAuditCall:      false,
			expectedAuditMetrics: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockMetricsService)
			tt.setupMock(mockService)

			mockAuditObserver := new(MockAuditObserver)
			tt.setupAudit(mockAuditObserver)

			auditSubject := audit.NewAuditSubject()
			auditSubject.Attach(mockAuditObserver)

			router := setupEncryptedUpdatesTestRouter(mockService, auditSubject, "", cryptoKeyPath)

			req, err := http.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(tt.body))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.RemoteAddr = "127.0.0.1:8080"

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Status code mismatch")

			mockService.AssertExpectations(t)

			if tt.expectAuditCall {
				assert.Equal(t, 1, mockAuditObserver.NotifyCallCount, "Audit should be called once")
			} else {
				assert.Equal(t, 0, mockAuditObserver.NotifyCallCount, "Audit should not be called")
			}
		})
	}
}

func TestUpdateJSONWithoutEncryption(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name                 string
		body                 string
		setupMock            func(*MockMetricsService)
		setupAudit           func(*MockAuditObserver)
		expectedStatus       int
		expectAuditCall      bool
		expectedAuditMetrics []string
	}{
		{
			name: "successful gauge update without encryption",
			body: `{"id":"temperature","type":"gauge","value":23.5}`,
			setupMock: func(m *MockMetricsService) {
				m.On("UpdateGauge", ctx, nil, "temperature", 23.5).Return(nil)
			},
			setupAudit: func(m *MockAuditObserver) {
				m.Err = nil
			},
			expectedStatus:       http.StatusOK,
			expectAuditCall:      true,
			expectedAuditMetrics: []string{"temperature"},
		},
		{
			name: "successful counter update without encryption",
			body: `{"id":"requests","type":"counter","delta":100}`,
			setupMock: func(m *MockMetricsService) {
				m.On("UpdateCounter", ctx, nil, "requests", int64(100)).Return(nil)
			},
			setupAudit: func(m *MockAuditObserver) {
				m.Err = nil
			},
			expectedStatus:       http.StatusOK,
			expectAuditCall:      true,
			expectedAuditMetrics: []string{"requests"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockMetricsService)
			tt.setupMock(mockService)

			mockAuditObserver := new(MockAuditObserver)
			tt.setupAudit(mockAuditObserver)

			auditSubject := audit.NewAuditSubject()
			auditSubject.Attach(mockAuditObserver)

			router := setupEncryptedUpdateJSONTestRouter(mockService, auditSubject, "", "")

			req, err := http.NewRequest(http.MethodPost, "/update/", bytes.NewBufferString(tt.body))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.RemoteAddr = "127.0.0.1:8080"

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Status code mismatch")

			mockService.AssertExpectations(t)

			if tt.expectAuditCall {
				assert.Equal(t, 1, mockAuditObserver.NotifyCallCount, "Audit should be called once")
				assert.Equal(t, tt.expectedAuditMetrics, mockAuditObserver.LastEvent.Metrics, "Audit metrics mismatch")
				assert.Equal(t, "127.0.0.1", mockAuditObserver.LastEvent.IPAddress, "Audit IP address mismatch")
				assert.Greater(t, mockAuditObserver.LastEvent.Timestamp, int64(0), "Audit timestamp should be set")
			} else {
				assert.Equal(t, 0, mockAuditObserver.NotifyCallCount, "Audit should not be called")
			}
		})
	}
}

func TestAuditErrorHandling(t *testing.T) {
	ctx := context.Background()
	mockService := new(MockMetricsService)
	mockService.On("UpdateGauge", ctx, nil, "temperature", 23.5).Return(nil)

	mockAuditObserver := &MockAuditObserver{
		Err: errors.New("audit error"),
	}

	auditSubject := audit.NewAuditSubject()
	auditSubject.Attach(mockAuditObserver)

	router := setupUpdateTestRouter(mockService, auditSubject)

	req, err := http.NewRequest(http.MethodPost, "/update/gauge/temperature/23.5", nil)
	assert.NoError(t, err)
	req.RemoteAddr = "127.0.0.1:8080"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))

	assert.Equal(t, 1, mockAuditObserver.NotifyCallCount)
	assert.Equal(t, []string{"temperature"}, mockAuditObserver.LastEvent.Metrics)
	assert.Equal(t, "127.0.0.1", mockAuditObserver.LastEvent.IPAddress)

	mockService.AssertExpectations(t)
}

func setupRouterWithTrustedSubnet(trustedSubnet string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := new(MockMetricsService)
	mockService.On("UpdateGauge", context.Background(), nil, "test", float64(123.45)).Return(nil)
	mockService.On("UpdateCounter", context.Background(), nil, "test", int64(0)).Return(nil)
	mockService.On("Updates", context.Background(), mock.Anything).Return(nil)
	auditSubject := audit.NewAuditSubject()

	handler := &metricsHandler{
		metricsService: mockService,
		trustedSubnet:  trustedSubnet,
		audit:          auditSubject,
	}

	router.POST("/update/", handler.UpdateJSON)
	router.POST("/updates/", handler.Updates)
	router.POST("/update/:type/:name/:value", handler.Update)

	return router
}

func TestTrustedSubnetValidation(t *testing.T) {
	tests := []struct {
		name           string
		trustedSubnet  string
		xRealIP        string
		expectedStatus int
	}{
		{
			name:           "empty trusted subnet - allows all IPs",
			trustedSubnet:  "",
			xRealIP:        "192.168.1.100",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "IP in trusted subnet /24",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "192.168.1.50",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "IP in trusted subnet /24 - edge case",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "192.168.1.255",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "IP not in trusted subnet /24",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "192.168.2.100",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "IP in trusted subnet /16",
			trustedSubnet:  "10.0.0.0/16",
			xRealIP:        "10.0.5.50",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "IP not in trusted subnet /16",
			trustedSubnet:  "10.0.0.0/16",
			xRealIP:        "10.1.0.50",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "single IP in trusted subnet",
			trustedSubnet:  "192.168.1.100/32",
			xRealIP:        "192.168.1.100",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "different single IP not in trusted subnet",
			trustedSubnet:  "192.168.1.100/32",
			xRealIP:        "192.168.1.101",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "localhost in trusted subnet",
			trustedSubnet:  "127.0.0.0/8",
			xRealIP:        "127.0.0.1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "localhost not in trusted subnet",
			trustedSubnet:  "10.0.0.0/8",
			xRealIP:        "127.0.0.1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "external IP not in private subnet",
			trustedSubnet:  "192.168.0.0/16",
			xRealIP:        "8.8.8.8",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "IPv4 in IPv6-like subnet fails",
			trustedSubnet:  "::1/128",
			xRealIP:        "127.0.0.1",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupRouterWithTrustedSubnet(tt.trustedSubnet)

			body := `{"id":"test","type":"gauge","value":123.45}`
			req, err := http.NewRequest(http.MethodPost, "/update/", bytes.NewBufferString(body))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.RemoteAddr = "127.0.0.1:8080"
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Status code mismatch for trustedSubnet=%s, xRealIP=%s", tt.trustedSubnet, tt.xRealIP)
		})
	}
}

func TestTrustedSubnetValidationUpdateEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		trustedSubnet  string
		xRealIP        string
		expectedStatus int
	}{
		{
			name:           "IP not in subnet returns 403",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "10.0.0.1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "IP in subnet proceeds to normal handling",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "192.168.1.100",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty subnet allows request to proceed",
			trustedSubnet:  "",
			xRealIP:        "10.0.0.1",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupRouterWithTrustedSubnet(tt.trustedSubnet)

			req, err := http.NewRequest(http.MethodPost, "/update/gauge/test/123.45", nil)
			assert.NoError(t, err)
			req.RemoteAddr = "127.0.0.1:8080"
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Status code mismatch")
		})
	}
}

func TestTrustedSubnetValidationUpdatesEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		trustedSubnet  string
		xRealIP        string
		expectedStatus int
	}{
		{
			name:           "IP not in subnet returns 403",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "172.16.0.1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "IP in subnet proceeds to normal handling",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "192.168.1.50",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupRouterWithTrustedSubnet(tt.trustedSubnet)

			body := `[{"id":"test","type":"gauge","value":123.45}]`
			req, err := http.NewRequest(http.MethodPost, "/updates/", bytes.NewBufferString(body))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.RemoteAddr = "127.0.0.1:8080"
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Status code mismatch")
		})
	}
}

func TestTrustedSubnetFallbackToRemoteAddr(t *testing.T) {
	tests := []struct {
		name           string
		trustedSubnet  string
		remoteAddr     string
		expectedStatus int
	}{
		{
			name:           "uses RemoteAddr when X-Real-IP not set and IP is valid",
			trustedSubnet:  "192.168.1.0/24",
			remoteAddr:     "192.168.1.100:8080",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "rejects when RemoteAddr not in subnet and no X-Real-IP",
			trustedSubnet:  "10.0.0.0/8",
			remoteAddr:     "192.168.1.100:8080",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupRouterWithTrustedSubnet(tt.trustedSubnet)

			body := `{"id":"test","type":"gauge","value":123.45}`
			req, err := http.NewRequest(http.MethodPost, "/update/", bytes.NewBufferString(body))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.RemoteAddr = tt.remoteAddr

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Status code mismatch")
		})
	}
}

func TestTrustedSubnetValidationResponseBody(t *testing.T) {
	router := setupRouterWithTrustedSubnet("192.168.1.0/24")

	body := `{"id":"test","type":"gauge","value":123.45}`
	req, err := http.NewRequest(http.MethodPost, "/update/", bytes.NewBufferString(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Real-IP", "10.0.0.1")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "forbidden")
}

func TestTrustedSubnetWithInvalidCIDR(t *testing.T) {
	router := setupRouterWithTrustedSubnet("invalid-cidr")

	body := `{"id":"test","type":"gauge","value":123.45}`
	req, err := http.NewRequest(http.MethodPost, "/update/", bytes.NewBufferString(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Real-IP", "192.168.1.100")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestValidateIPMethod(t *testing.T) {
	tests := []struct {
		name          string
		trustedSubnet string
		clientIP      string
		expected      bool
	}{
		{
			name:          "empty subnet - all IPs allowed",
			trustedSubnet: "",
			clientIP:      "192.168.1.100",
			expected:      true,
		},
		{
			name:          "IP in subnet",
			trustedSubnet: "192.168.1.0/24",
			clientIP:      "192.168.1.50",
			expected:      true,
		},
		{
			name:          "IP not in subnet",
			trustedSubnet: "192.168.1.0/24",
			clientIP:      "10.0.0.1",
			expected:      false,
		},
		{
			name:          "exact IP match /32",
			trustedSubnet: "192.168.1.100/32",
			clientIP:      "192.168.1.100",
			expected:      true,
		},
		{
			name:          "exact IP mismatch /32",
			trustedSubnet: "192.168.1.100/32",
			clientIP:      "192.168.1.101",
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &metricsHandler{
				trustedSubnet: tt.trustedSubnet,
			}

			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request, _ = http.NewRequest("POST", "/", nil)
			c.Request.RemoteAddr = "127.0.0.1:8080"
			if tt.clientIP != "" {
				c.Request.Header.Set("X-Real-IP", tt.clientIP)
			}

			result := handler.validateIP(c)
			if tt.expected {
				assert.True(t, result, "Expected IP to be valid")
			} else {
				assert.False(t, result, "Expected IP to be invalid")
			}
		})
	}
}
