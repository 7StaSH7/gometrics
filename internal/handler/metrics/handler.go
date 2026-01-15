// Package metrics provides HTTP handlers for metrics operations.
package metrics

import (
	"context"
	"net"
	"net/http"

	"github.com/7StaSH7/gometrics/internal/audit"
	"github.com/7StaSH7/gometrics/internal/logger"
	"github.com/7StaSH7/gometrics/internal/service/metrics"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type metricsHandler struct {
	metricsService metrics.MetricsService
	hashKey        string
	cryptoKey      string
	audit          *audit.AuditSubject
	trustedSubnet  string
}

// MetricsHandler defines the interface for handling metric-related HTTP requests,
// including updating, retrieving, and registering routes.
type MetricsHandler interface {
	UpdateJSON(*gin.Context)
	GetJSON(*gin.Context)
	Updates(*gin.Context)

	Update(*gin.Context)
	GetOne(*gin.Context)

	Register(*gin.Engine)

	GetMany(*gin.Context)
}

// New creates a new MetricsHandler with the given metrics service, hash key, crypto key, and audit subject.
func New(s metrics.MetricsService, key string, cryptoKey string, asub *audit.AuditSubject, trustedSubnet string) MetricsHandler {
	return &metricsHandler{
		metricsService: s,
		hashKey:        key,
		cryptoKey:      cryptoKey,
		audit:          asub,
		trustedSubnet:  trustedSubnet,
	}
}

// auditEvent creates and sends an audit event if audit is enabled.
func (h *metricsHandler) auditEvent(ctx context.Context, metrics []string, ip string) {
	if h.audit != nil {
		event := audit.CreateAuditEvent(metrics, ip)
		if err := h.audit.NotifyAll(ctx, event); err != nil {
			logger.Log.Error("Audit notification failed", zap.Error(err))
		}
	}
}
// validateIP parses and validates X-Real-IP header to be CIDR format.
func (h *metricsHandler) validateIP(c *gin.Context) bool {
	if h.trustedSubnet == "" {
		return true
	}

	clientIP := c.GetHeader("X-Real-IP")
	if clientIP == "" {
		clientIP = c.ClientIP()
	}

	_, ipNet, err := net.ParseCIDR(h.trustedSubnet)
	if err != nil {
		logger.Log.Error("invalid trusted subnet", zap.Error(err))
		return false
	}

	ip := net.ParseIP(clientIP)
	if ip == nil || !ipNet.Contains(ip) {
		logger.Log.Warn("IP not in trusted subnet", zap.String("ip", clientIP), zap.String("subnet", h.trustedSubnet))
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return false
	}

	return true
}

// Register registers the metric routes with the Gin engine.
func (h *metricsHandler) Register(e *gin.Engine) {
	e.POST("/update/:type/:name/:value", h.Update)
	e.GET("/value/:type/:name", h.GetOne)

	e.POST("/update/", h.UpdateJSON)
	e.POST("/value/", h.GetJSON)
	e.POST("/updates/", h.Updates)

	e.GET("", h.GetMany)
}
