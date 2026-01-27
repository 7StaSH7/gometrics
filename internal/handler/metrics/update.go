// Package metrics provides HTTP handlers for metrics operations.
package metrics

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/7StaSH7/gometrics/internal/logger"
	"github.com/7StaSH7/gometrics/internal/model"
	"github.com/7StaSH7/gometrics/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UpdateMetricInput represents the input parameters for updating a metric via URI.
type UpdateMetricInput struct {
	MType string `uri:"type"`
	Name  string `uri:"name"`
	Value string `uri:"value"`
}

// Update handles POST requests to update a single metric by type, name, and value.
func (h *metricsHandler) Update(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")

	if !h.validateIP(c) {
		return
	}

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

		if err := h.metricsService.UpdateGauge(c.Request.Context(), nil, input.Name, parsedValue); err != nil {
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

		if err := h.metricsService.UpdateCounter(c.Request.Context(), nil, input.Name, parsedValue); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}

	h.auditEvent(c.Request.Context(), []string{input.Name}, c.ClientIP())
	c.Status(http.StatusOK)
}

// decryptRequestBody reads and optionally decrypts the request body.
// If cryptoKey is configured, it decrypts the body using RSA OAEP.
// Otherwise, it returns the raw body data.
func (h *metricsHandler) decryptRequestBody(c *gin.Context) ([]byte, error) {
	if h.cryptoKey == "" {
		return io.ReadAll(c.Request.Body)
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Log.Debug("cannot read request body", zap.Error(err))
		return nil, err
	}

	decrypted, err := utils.Decrypt(body)
	if err != nil {
		logger.Log.Debug("cannot decrypt request body", zap.Error(err))
		return nil, err
	}

	return decrypted, nil
}

func (h *metricsHandler) UpdateJSON(c *gin.Context) {
	if !h.validateIP(c) {
		return
	}

	var hash string
	if h.hashKey != "" {
		hash = c.GetHeader("HashSHA256")
	}

	var body model.Metrics

	decryptedBody, err := h.decryptRequestBody(c)
	if err != nil {
		if h.cryptoKey == "" {
			if err := c.ShouldBindJSON(&body); err != nil {
				logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
				c.JSON(http.StatusBadRequest, gin.H{"error": err})
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
	} else {
		if err := json.Unmarshal(decryptedBody, &body); err != nil {
			logger.Log.Debug("cannot unmarshal decrypted JSON body", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
	}

	logger.Log.Debug("decoded JSON body", zap.Any("body", body))

	var jsonData []byte
	var expectedHash string
	if h.hashKey != "" && hash != "" {
		if len(decryptedBody) > 0 {
			jsonData = decryptedBody
		} else {
			jsonData, err = json.Marshal(body)
			if err != nil {
				logger.Log.Debug("cannot marshal JSON body", zap.Error(err))
				c.JSON(http.StatusBadRequest, gin.H{"error": err})
				return
			}
		}

		expectedHash = utils.GenerateSHA256Bytes(jsonData, h.hashKey)

		if !utils.VerifySHA256(expectedHash, hash) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
			return
		}
	}

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
			if err := h.metricsService.UpdateCounter(c.Request.Context(), nil, body.ID, *body.Delta); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err})
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
			if err := h.metricsService.UpdateGauge(c.Request.Context(), nil, body.ID, *body.Value); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err})
				return
			}
		}
	}

	if h.hashKey != "" && expectedHash != "" {
		c.Header("HashSHA256", expectedHash)
	}

	h.auditEvent(c.Request.Context(), []string{body.ID}, c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// Updates handles POST requests to update multiple metrics in batch.
func (h *metricsHandler) Updates(c *gin.Context) {
	if !h.validateIP(c) {
		return
	}

	var hash string
	if h.hashKey != "" {
		hash = c.GetHeader("HashSHA256")
	}

	metrics := make([]model.Metrics, 0)

	decryptedBody, err := h.decryptRequestBody(c)
	if err != nil {
		if h.cryptoKey == "" {
			if err := c.ShouldBindJSON(&metrics); err != nil {
				logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
				c.JSON(http.StatusBadRequest, gin.H{"error": err})
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
	} else {
		if err := json.Unmarshal(decryptedBody, &metrics); err != nil {
			logger.Log.Debug("cannot unmarshal decrypted JSON body", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
	}

	var jsonData []byte
	var expectedHash string
	if h.hashKey != "" && hash != "" {
		if len(decryptedBody) > 0 {
			jsonData = decryptedBody
		} else {
			jsonData, err = json.Marshal(metrics)
			if err != nil {
				logger.Log.Debug("cannot marshal JSON body", zap.Error(err))
				c.JSON(http.StatusBadRequest, gin.H{"error": err})
				return
			}
		}

		expectedHash = utils.GenerateSHA256Bytes(jsonData, h.hashKey)

		if !utils.VerifySHA256(expectedHash, hash) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
			return
		}
	}

	for _, m := range metrics {
		if m.MType != model.Counter && m.MType != model.Gauge {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad type"})
			return
		}

		if m.ID == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "bad id"})
			return
		}

		switch m.MType {
		case model.Counter:
			if m.Delta == nil {
				logger.Log.Debug("'Delta' field is missing")
				c.JSON(http.StatusBadRequest, gin.H{"error": "'Delta' is missing"})
				return
			}

		case model.Gauge:
			if m.Value == nil {
				logger.Log.Debug("'Value' field is missing", zap.String("field", m.ID))
				c.JSON(http.StatusBadRequest, gin.H{"error": "'Value' is missing"})
				return
			}
		}
	}

	if err := h.metricsService.Updates(c.Request.Context(), metrics); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	if h.hashKey != "" && expectedHash != "" {
		c.Header("HashSHA256", expectedHash)
	}

	h.auditEvent(
		c.Request.Context(),
		func() []string {
			m := make([]string, 0, len(metrics))
			for _, met := range metrics {
				m = append(m, met.ID)
			}
			return m
		}(),
		c.ClientIP(),
	)

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
