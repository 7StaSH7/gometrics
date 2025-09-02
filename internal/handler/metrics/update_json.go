package metrics

import (
	"net/http"

	"github.com/7StaSH7/gometrics/internal/logger"
	"github.com/7StaSH7/gometrics/internal/model"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h *metricsHandler) UpdateJson(c *gin.Context) {
	var body model.Metrics
	if err := c.ShouldBindJSON(&body); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	logger.Log.Debug("decoded JSON body", zap.Any("body", &body))

	if body.MType != model.Counter && body.MType != model.Gauge {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if body.ID == "" {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	switch body.MType {
	case model.Counter:
		{
			if body.Delta == nil {
				logger.Log.Debug("'Delta' field is missing")
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			if err := h.metricsService.UpdateCounter(body.ID, *body.Delta); err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
		}
	case model.Gauge:
		{
			if body.Value == nil {
				logger.Log.Debug("'Value' field is missing")
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			if err := h.metricsService.UpdateGauge(body.ID, *body.Value); err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
		}
	}

	c.Header("Content-Type", "application/json")
	c.Status(http.StatusOK)
}
