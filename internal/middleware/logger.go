package middleware

import (
	"time"

	"github.com/7StaSH7/gometrics/internal/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RequestLogger() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		logger.Log.Info("request completed",
			zap.String("uri", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.Duration("duration", duration),
			zap.Any("response", gin.H{
				"status": c.Writer.Status(),
				"size":   c.Writer.Size(),
			},
			),
		)
	})
}
