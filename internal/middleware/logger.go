package middleware

import (
	"bytes"
	"time"

	"github.com/7StaSH7/gometrics/internal/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func RequestLogger(c *gin.Context) {
	// blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	// c.Writer = blw

	start := time.Now()
	c.Next()
	duration := time.Since(start)
	logger.Log.Info("request completed",
		zap.String("uri", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.Duration("duration", duration),
		// zap.Any("resp", gin.H{
		// 	"status": c.Writer.Status(),
		// 	"size":   c.Writer.Size(),
		// 	"body":   blw.body.String(),
		// },
		// ),
	)
}
