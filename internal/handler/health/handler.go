package health

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type healthHandler struct {
	pool *pgxpool.Pool
}

// HealthHandler defines the interface for health check handlers.
type HealthHandler interface {
	Register(*gin.Engine)
}

// New creates a new HealthHandler with the given database pool.
func New(pool *pgxpool.Pool) HealthHandler {
	return &healthHandler{
		pool: pool,
	}
}

// Register registers the health check route with the Gin engine.
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
