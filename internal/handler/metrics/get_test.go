package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/7StaSH7/gometrics/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupGetTestRouter(service *MockMetricsService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := &metricsHandler{
		metricsService: service,
	}

	router.GET("/value/:type/:name", handler.Get)

	return router
}

func (m *MockMetricsService) GetOne(mType, name string) string {
	args := m.Called(mType, name)

	return args.String(0)
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
				m.On("GetOne", model.Gauge, "temperature").Return("23.5")
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "23.5",
		},
		{
			name: "successful counter retrieval",
			url:  "/value/counter/requests",
			setupMock: func(m *MockMetricsService) {
				m.On("GetOne", model.Counter, "requests").Return("100")
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "100",
		},
		{
			name: "successful gauge retrieval with negative value",
			url:  "/value/gauge/temperature",
			setupMock: func(m *MockMetricsService) {
				m.On("GetOne", model.Gauge, "temperature").Return("-15.3")
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "-15.3",
		},
		{
			name: "successful gauge retrieval with zero value",
			url:  "/value/gauge/pressure",
			setupMock: func(m *MockMetricsService) {
				m.On("GetOne", model.Gauge, "pressure").Return("0")
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "0",
		},
		{
			name: "successful counter retrieval with zero value",
			url:  "/value/counter/errors",
			setupMock: func(m *MockMetricsService) {
				m.On("GetOne", model.Counter, "errors").Return("0")
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "0",
		},
		{
			name: "successful gauge retrieval with large value",
			url:  "/value/gauge/temperature",
			setupMock: func(m *MockMetricsService) {
				m.On("GetOne", model.Gauge, "temperature").Return("999999.999999")
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "999999.999999",
		},
		{
			name: "successful counter retrieval with large value",
			url:  "/value/counter/requests",
			setupMock: func(m *MockMetricsService) {
				m.On("GetOne", model.Counter, "requests").Return("9223372036854775807")
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
			name: "metric not found - gauge",
			url:  "/value/gauge/nonexistent",
			setupMock: func(m *MockMetricsService) {
				m.On("GetOne", model.Gauge, "nonexistent").Return("")
			},
			expectedStatus: http.StatusNotFound,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "",
		},
		{
			name: "metric not found - counter",
			url:  "/value/counter/nonexistent",
			setupMock: func(m *MockMetricsService) {
				m.On("GetOne", model.Counter, "nonexistent").Return("")
			},
			expectedStatus: http.StatusNotFound,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "",
		},
		{
			name: "gauge with scientific notation",
			url:  "/value/gauge/scientific",
			setupMock: func(m *MockMetricsService) {
				m.On("GetOne", model.Gauge, "scientific").Return("1.23e+10")
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "1.23e+10",
		},
		{
			name: "metric name with special characters",
			url:  "/value/gauge/cpu_usage_percent",
			setupMock: func(m *MockMetricsService) {
				m.On("GetOne", model.Gauge, "cpu_usage_percent").Return("75.5")
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "75.5",
		},
		{
			name: "metric name with numbers",
			url:  "/value/counter/http_200_responses",
			setupMock: func(m *MockMetricsService) {
				m.On("GetOne", model.Counter, "http_200_responses").Return("1500")
			},
			expectedStatus: http.StatusOK,
			expectedHeader: "text/plain; charset=utf-8",
			expectedBody:   "1500",
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
