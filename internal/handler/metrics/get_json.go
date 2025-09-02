package metrics

import (
	"encoding/json"
	"net/http"

	"github.com/7StaSH7/gometrics/internal/model"
	"github.com/gin-gonic/gin"
)

func (h *metricsHandler) GetJson(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	var body model.Metrics

	dec := json.NewDecoder(c.Request.Body)
	if err := dec.Decode(&body); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if body.ID == "" {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if body.MType != model.Counter && body.MType != model.Gauge {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	switch body.MType {
	case model.Counter:
		{
			value := h.metricsService.GetCounter(body.ID)
			if value == 0 {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}

			body.Delta = &value
		}
	case model.Gauge:
		{
			value := h.metricsService.GetGauge(body.ID)
			if value == 0 {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}

			body.Value = &value
		}
	}

	c.JSON(http.StatusOK, body)
}
