package metrics

import (
	"github.com/7StaSH7/gometrics/internal/service/metrics"
	"github.com/gin-gonic/gin"
)

type metricsHandler struct {
	metricsService metrics.MetricsService
}

type MetricsHandler interface {
	UpdateJson(*gin.Context)
	GetJson(*gin.Context)

	Update(*gin.Context)
	GetOne(*gin.Context)

	Register(*gin.Engine)

	GetMany(*gin.Context)
}

func NewHandler(s metrics.MetricsService) MetricsHandler {
	return &metricsHandler{
		metricsService: s,
	}
}

func (h *metricsHandler) Register(e *gin.Engine) {
	e.POST("/update/:type/:name/:value", h.Update)
	e.GET("/value/:type/:name", h.GetOne)

	e.POST("/update/", h.UpdateJson)
	e.POST("/value/", h.GetJson)

	e.GET("", h.GetMany)
}
