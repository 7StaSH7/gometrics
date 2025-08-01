package metrics

import (
	"net/http"
	"strconv"

	"github.com/7StaSH7/gometrics/internal/model"
	"github.com/gin-gonic/gin"
)

type MetricInput struct {
	MType string `uri:"type"`
	Name  string `uri:"name"`
	Value string `uri:"value"`
}

func (h *metricsHandler) Update(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")

	var metric MetricInput
	if err := c.ShouldBindUri(&metric); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	if metric.Name == "" {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if metric.MType != model.Counter && metric.MType != model.Gauge {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if metric.MType == model.Gauge {
		parsedValue, err := strconv.ParseFloat(metric.Value, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if err := h.metricsService.UpdateMetric(metric.MType, metric.Name, parsedValue); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}

	if metric.MType == model.Counter {
		parsedValue, err := strconv.ParseInt(metric.Value, 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if err := h.metricsService.UpdateMetric(metric.MType, metric.Name, parsedValue); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}

	c.Status(http.StatusOK)
}
