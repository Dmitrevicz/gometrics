package server

import (
	"time"

	"github.com/Dmitrevicz/gometrics/internal/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestLogger — middleware-logger for incoming HTTP-requests.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		ts := time.Now()

		c.Next()

		logger.Log.Info("got HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("code", c.Writer.Status()),
			zap.Int("size", c.Writer.Size()),
			zap.Duration("duration", time.Since(ts)),
		)
	}
}
