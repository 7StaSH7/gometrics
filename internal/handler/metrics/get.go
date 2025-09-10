package metrics

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/7StaSH7/gometrics/internal/logger"
	"github.com/7StaSH7/gometrics/internal/model"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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

			c.String(http.StatusOK, fmt.Sprintf("%v", value))
			return
		}
	case model.Gauge:
		{
			value, err := h.metricsService.GetGauge(input.Name)
			if err != nil {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}

			c.String(http.StatusOK, fmt.Sprintf("%v", value))
			return
		}
	}
}

func (h *metricsHandler) GetJSON(c *gin.Context) {
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
			value, err := h.metricsService.GetCounter(body.ID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "metric not found"})
				return
			}

			body.Delta = &value
		}
	case model.Gauge:
		{
			value, err := h.metricsService.GetGauge(body.ID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "metric not found"})
				return
			}

			body.Value = &value
		}
	}

	c.JSON(http.StatusOK, body)
}

func (h *metricsHandler) GetMany(c *gin.Context) {
	metrics := h.metricsService.GetMany()

	c.HTML(http.StatusOK, "metrics.tmpl", gin.H{
		"metrics": metrics,
	})
}
