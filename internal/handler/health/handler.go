package health

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type healthHandler struct {
	conn *pgx.Conn
}

type HealthHandler interface {
	Register(*gin.Engine)
}

func New(conn *pgx.Conn) HealthHandler {
	return &healthHandler{
		conn: conn,
	}
}

func (h *healthHandler) Register(e *gin.Engine) {
	e.GET("/ping", func(c *gin.Context) {
		if err := h.conn.Ping(c); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"status": "OK"})
	})
}
