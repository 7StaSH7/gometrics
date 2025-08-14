package metrics

import (
	"net/http"

	"github.com/7StaSH7/gometrics/internal/model"
	"github.com/gin-gonic/gin"
)

type GetMetricInput struct {
	MType string `uri:"type"`
	Name  string `uri:"name"`
}

func (h *metricsHandler) GetOne(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")

	var input GetMetricInput
	if err := c.ShouldBindUri(&input); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	if input.Name == "" {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if input.MType != model.Counter && input.MType != model.Gauge {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	value := h.metricsService.GetOne(input.MType, input.Name)
	if value == "" {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.String(http.StatusOK, value)
}
