package metrics

import (
	"fmt"
	"net/http"
	"strconv"

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

	switch input.MType {
	case model.Counter:
		{
			value, err := h.metricsService.GetCounter(input.Name)
			if err != nil {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}

			c.String(http.StatusOK, fmt.Sprintf("%d", value))
			return
		}
	case model.Gauge:
		{
			value, err := h.metricsService.GetGauge(input.Name)
			if err != nil {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}

			c.String(http.StatusOK, strconv.FormatFloat(value, 'f', -1, 64))
			return
		}
	}
}
