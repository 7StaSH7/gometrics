package metrics

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *metricsHandler) Metrics(c *gin.Context) {
	metrics := h.metricsService.GetMany()
	
	c.HTML(http.StatusOK, "metrics.tmpl", gin.H{
		"metrics": metrics,
	})
}
