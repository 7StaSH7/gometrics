package metrics

import (
	"net/http"
	"strconv"

	"github.com/7StaSH7/gometrics/internal/model"
	"github.com/gin-gonic/gin"
)

type UpdateMetricInput struct {
	MType string `uri:"type"`
	Name  string `uri:"name"`
	Value string `uri:"value"`
}

func (h *metricsHandler) Update(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")

	var input UpdateMetricInput
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

	if input.MType == model.Gauge {
		parsedValue, err := strconv.ParseFloat(input.Value, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if err := h.metricsService.Update(input.MType, input.Name, parsedValue); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}

	if input.MType == model.Counter {
		parsedValue, err := strconv.ParseInt(input.Value, 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if err := h.metricsService.Update(input.MType, input.Name, parsedValue); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}

	c.Status(http.StatusOK)
}
