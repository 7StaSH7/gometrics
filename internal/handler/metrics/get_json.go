package metrics

import (
	"encoding/json"
	"net/http"

	"github.com/7StaSH7/gometrics/internal/logger"
	"github.com/7StaSH7/gometrics/internal/model"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h *metricsHandler) GetJson(c *gin.Context) {
	var body model.Metrics

	dec := json.NewDecoder(c.Request.Body)
	if err := dec.Decode(&body); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "bad id"})
		return
	}

	if body.MType != model.Counter && body.MType != model.Gauge {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad type"})
		return
	}

	switch body.MType {
	case model.Counter:
		{
			value := h.metricsService.GetCounter(body.ID)
			if value == 0 {
				c.JSON(http.StatusNotFound, gin.H{"error": "metric not found"})
				return
			}

			body.Delta = &value
		}
	case model.Gauge:
		{
			value := h.metricsService.GetGauge(body.ID)
			if value == 0 {
				c.JSON(http.StatusNotFound, gin.H{"error": "metric not found"})
				return
			}

			body.Value = &value
		}
	}

	c.JSON(http.StatusOK, body)
}
