package middleware

import (
	"log/slog"
	"time"

	"github.com/ali/hesab-keepnet/backend/internal/requestid"
	"github.com/gin-gonic/gin"
)

func Logger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		status := c.Writer.Status()
		attrs := []any{
			"method", c.Request.Method,
			"path", path,
			"route", c.FullPath(),
			"status", status,
			"duration_ms", float64(time.Since(start).Microseconds()) / 1000.0,
			"request_id", requestid.FromContext(c),
			"ip", c.ClientIP(),
			"bytes", c.Writer.Size(),
		}
		if query != "" {
			attrs = append(attrs, "query", query)
		}

		switch {
		case status >= 500:
			log.Error("http_request", attrs...)
		case status >= 400:
			log.Warn("http_request", attrs...)
		default:
			log.Info("http_request", attrs...)
		}
	}
}
