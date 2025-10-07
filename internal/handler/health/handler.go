package health

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type healthHandler struct {
	pool *pgxpool.Pool
}

type HealthHandler interface {
	Register(*gin.Engine)
}

func New(pool *pgxpool.Pool) HealthHandler {
	return &healthHandler{
		pool: pool,
	}
}

func (h *healthHandler) Register(e *gin.Engine) {
	e.GET("/ping", func(c *gin.Context) {
		if h.pool == nil {
			c.JSON(500, gin.H{"error": "connection is nil"})
			return
		}
		if err := h.pool.Ping(c); err != nil {
			c.JSON(500, gin.H{"error": err})
			return
		}

		c.JSON(200, gin.H{"status": "OK"})
	})
}
