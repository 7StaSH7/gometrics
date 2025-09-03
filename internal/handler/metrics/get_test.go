package metrics

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupGetTestRouter(service *MockMetricsService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := &metricsHandler{
		metricsService: service,
	}

	router.GET("/value/:type/:name", handler.GetOne)

	return router
}

func (m *MockMetricsService) GetCounter(name string) (int64, error) {
	args := m.Called(name)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockMetricsService) GetGauge(name string) (float64, error) {
	args := m.Called(name)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockMetricsService) GetMany() map[string]string {
	args := m.Called()

	return args.Get(0).(map[string]string)
}

func TestGet(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		setupMock      func(*MockMetricsService)
		expectedStatus int
		expectedHeader string
		expectedBody   string
	}{
		{
			name: "successful gauge retrieval",
			url:  "/value/gauge/temperature",
			setupMock: func(m *MockMetricsService) {
				m.On("GetGauge", "temperature").Return(float64(23.5), nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "23.5",
		},
		{
			name: "successful counter retrieval",
			url:  "/value/counter/requests",
			setupMock: func(m *MockMetricsService) {
				m.On("GetCounter", "requests").Return(int64(100), nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "100",
		},
		{
			name: "successful gauge retrieval with negative value",
			url:  "/value/gauge/temperature",
			setupMock: func(m *MockMetricsService) {
				m.On("GetGauge", "temperature").Return(float64(-15.3), nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "-15.3",
		},
		{
			name: "gauge with zero value returns 200",
			url:  "/value/gauge/pressure",
			setupMock: func(m *MockMetricsService) {
				m.On("GetGauge", "pressure").Return(float64(0), nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "0",
		},
		{
			name: "counter with zero value returns 200",
			url:  "/value/counter/errors",
			setupMock: func(m *MockMetricsService) {
				m.On("GetCounter", "errors").Return(int64(0), nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "0",
		},
		{
			name: "successful gauge retrieval with large value",
			url:  "/value/gauge/temperature",
			setupMock: func(m *MockMetricsService) {
				m.On("GetGauge", "temperature").Return(float64(999999.999999), nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "999999.999999",
		},
		{
			name: "successful counter retrieval with large value",
			url:  "/value/counter/requests",
			setupMock: func(m *MockMetricsService) {
				m.On("GetCounter", "requests").Return(int64(9223372036854775807), nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "9223372036854775807",
		},
		{
			name:           "invalid metric type",
			url:            "/value/histogram/invalid",
			setupMock:      func(m *MockMetricsService) {},
			expectedStatus: http.StatusBadRequest,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "",
		},
		{
			name: "gauge with scientific notation value",
			url:  "/value/gauge/scientific",
			setupMock: func(m *MockMetricsService) {
				m.On("GetGauge", "scientific").Return(float64(1.23e+10), nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "1.23e+10",
		},
		{
			name: "metric name with special characters",
			url:  "/value/gauge/cpu_usage_percent",
			setupMock: func(m *MockMetricsService) {
				m.On("GetGauge", "cpu_usage_percent").Return(float64(75.5), nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "75.5",
		},
		{
			name: "metric name with numbers",
			url:  "/value/counter/http_200_responses",
			setupMock: func(m *MockMetricsService) {
				m.On("GetCounter", "http_200_responses").Return(int64(1500), nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "1500",
		},
		{
			name: "gauge with very small positive value",
			url:  "/value/gauge/small_value",
			setupMock: func(m *MockMetricsService) {
				m.On("GetGauge", "small_value").Return(float64(0.001), nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "0.001",
		},
		{
			name: "gauge with very small negative value",
			url:  "/value/gauge/small_negative",
			setupMock: func(m *MockMetricsService) {
				m.On("GetGauge", "small_negative").Return(float64(-0.001), nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "-0.001",
		},
		{
			name: "counter with value 1",
			url:  "/value/counter/single_request",
			setupMock: func(m *MockMetricsService) {
				m.On("GetCounter", "single_request").Return(int64(1), nil)
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "1",
		},
		{
			name: "metric not found - gauge",
			url:  "/value/gauge/nonexistent",
			setupMock: func(m *MockMetricsService) {
				m.On("GetGauge", "nonexistent").Return(float64(0), fmt.Errorf("gauge metric 'nonexistent' not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "",
		},
		{
			name: "metric not found - counter",
			url:  "/value/counter/nonexistent",
			setupMock: func(m *MockMetricsService) {
				m.On("GetCounter", "nonexistent").Return(int64(0), fmt.Errorf("counter metric 'nonexistent' not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockMetricsService)
			tt.setupMock(mockService)

			router := setupGetTestRouter(mockService)

			req, err := http.NewRequest(http.MethodGet, tt.url, nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Status code mismatch")
			assert.Equal(t, tt.expectedHeader, w.Header().Get("Content-Type"), "Content-Type header mismatch")
			assert.Equal(t, tt.expectedBody, w.Body.String(), "Response body mismatch")

			mockService.AssertExpectations(t)
		})
	}
}
