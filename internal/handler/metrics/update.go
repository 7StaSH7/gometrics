package metrics

import (
	"net/http"
	"strconv"

	"github.com/7StaSH7/gometrics/internal/logger"
	"github.com/7StaSH7/gometrics/internal/model"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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

		if err := h.metricsService.UpdateGauge(input.Name, parsedValue); err != nil {
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

		if err := h.metricsService.UpdateCounter(input.Name, parsedValue); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}

	c.Status(http.StatusOK)
}

func (h *metricsHandler) UpdateJSON(c *gin.Context) {
	var body model.Metrics
	if err := c.ShouldBindJSON(&body); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Log.Debug("decoded JSON body", zap.Any("body", body))

	if body.MType != model.Counter && body.MType != model.Gauge {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad type"})
		return
	}

	if body.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "bad id"})
		return
	}

	switch body.MType {
	case model.Counter:
		{
			if body.Delta == nil {
				logger.Log.Debug("'Delta' field is missing")
				c.JSON(http.StatusBadRequest, gin.H{"error": "'Delta' is missing"})
				return
			}
			if err := h.metricsService.UpdateCounter(body.ID, *body.Delta); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}
	case model.Gauge:
		{
			if body.Value == nil {
				logger.Log.Debug("'Value' field is missing")
				c.JSON(http.StatusBadRequest, gin.H{"error": "'Value' is missing"})
				return
			}
			if err := h.metricsService.UpdateGauge(body.ID, *body.Value); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{})
}
