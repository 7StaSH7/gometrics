package metrics

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/7StaSH7/gometrics/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMetricsService struct {
	mock.Mock
}

func (m *MockMetricsService) Update(mType, name string, value any) error {
	args := m.Called(mType, name, value)

	return args.Error(0)
}

func (m *MockMetricsService) GetOne(mType, name string) string {
	args := m.Called(mType, name)

	return args.String(0)
}

func setupTestRouter(service *MockMetricsService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := &metricsHandler{
		metricsService: service,
	}

	router.POST("/update/:type/:name/:value", handler.Update)

	return router
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		setupMock      func(*MockMetricsService)
		expectedStatus int
		expectedHeader string
	}{
		{
			name: "successful gauge update",
			url:  "/update/gauge/temperature/23.5",
			setupMock: func(m *MockMetricsService) {
				m.On("Update", model.Gauge, "temperature", 23.5).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
		},
		{
			name: "successful counter update",
			url:  "/update/counter/requests/100",
			setupMock: func(m *MockMetricsService) {
				m.On("Update", model.Counter, "requests", int64(100)).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
		},
		{
			name: "successful gauge update with negative value",
			url:  "/update/gauge/temperature/-15.3",
			setupMock: func(m *MockMetricsService) {
				m.On("Update", model.Gauge, "temperature", -15.3).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
		},
		{
			name: "successful gauge update with zero value",
			url:  "/update/gauge/pressure/0.0",
			setupMock: func(m *MockMetricsService) {
				m.On("Update", model.Gauge, "pressure", 0.0).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
		},
		{
			name: "successful counter update with zero value",
			url:  "/update/counter/errors/0",
			setupMock: func(m *MockMetricsService) {
				m.On("Update", model.Counter, "errors", int64(0)).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
		},
		{
			name:           "invalid metric type",
			url:            "/update/histogram/invalid/123",
			setupMock:      func(m *MockMetricsService) {},
			expectedStatus: http.StatusBadRequest,
			expectedHeader: "text/plain; charset=utf-8",
		},
		{
			name:           "empty metric name",
			url:            "/update/gauge//123.45",
			setupMock:      func(m *MockMetricsService) {},
			expectedStatus: http.StatusNotFound,
			expectedHeader: "text/plain; charset=utf-8",
		},
		{
			name:           "invalid gauge value - not a float",
			url:            "/update/gauge/temperature/invalid",
			setupMock:      func(m *MockMetricsService) {},
			expectedStatus: http.StatusBadRequest,
			expectedHeader: "text/plain; charset=utf-8",
		},
		{
			name:           "invalid counter value - not an integer",
			url:            "/update/counter/requests/12.34",
			setupMock:      func(m *MockMetricsService) {},
			expectedStatus: http.StatusBadRequest,
			expectedHeader: "text/plain; charset=utf-8",
		},
		{
			name:           "invalid counter value - not a number",
			url:            "/update/counter/requests/abc",
			setupMock:      func(m *MockMetricsService) {},
			expectedStatus: http.StatusBadRequest,
			expectedHeader: "text/plain; charset=utf-8",
		},
		{
			name: "service error for gauge update",
			url:  "/update/gauge/temperature/25.0",
			setupMock: func(m *MockMetricsService) {
				m.On("Update", model.Gauge, "temperature", 25.0).Return(errors.New("service error"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedHeader: "text/plain; charset=utf-8",
		},
		{
			name: "service error for counter update",
			url:  "/update/counter/requests/50",
			setupMock: func(m *MockMetricsService) {
				m.On("Update", model.Counter, "requests", int64(50)).Return(errors.New("service error"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedHeader: "text/plain; charset=utf-8",
		},
		{
			name: "large gauge value",
			url:  "/update/gauge/temperature/999999.999999",
			setupMock: func(m *MockMetricsService) {
				m.On("Update", model.Gauge, "temperature", 999999.999999).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
		},
		{
			name: "large counter value",
			url:  "/update/counter/requests/9223372036854775807",
			setupMock: func(m *MockMetricsService) {
				m.On("Update", model.Counter, "requests", int64(9223372036854775807)).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
		},
		{
			name:           "counter value overflow",
			url:            "/update/counter/requests/99999999999999999999",
			setupMock:      func(m *MockMetricsService) {},
			expectedStatus: http.StatusBadRequest,
			expectedHeader: "text/plain; charset=utf-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockMetricsService)
			tt.setupMock(mockService)

			router := setupTestRouter(mockService)

			req, err := http.NewRequest(http.MethodPost, tt.url, nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Status code mismatch")
			assert.Equal(t, tt.expectedHeader, w.Header().Get("Content-Type"), "Content-Type header mismatch")

			mockService.AssertExpectations(t)
		})
	}
}
