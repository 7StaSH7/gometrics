package metrics

import (
	"github.com/7StaSH7/gometrics/internal/service/metrics"
	"github.com/gin-gonic/gin"
)

type metricsHandler struct {
	metricsService metrics.MetricsService
}

type MetricsHandler interface {
	Update(*gin.Context)
	Get(*gin.Context)
	Register(*gin.Engine)
	Metrics(*gin.Context)
}

func NewHandler(s metrics.MetricsService) MetricsHandler {
	return &metricsHandler{
		metricsService: s,
	}
}

func (h *metricsHandler) Register(e *gin.Engine) {
	e.POST("/update/:type/:name/:value", h.Update)
	e.GET("/value/:type/:name", h.Get)
	e.GET("", h.Metrics)
}
