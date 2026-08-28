package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/ali/hesab-keepnet/backend/internal/httpx"
	"github.com/ali/hesab-keepnet/backend/internal/requestid"
	"github.com/gin-gonic/gin"
)

func Recovery(appEnv string, log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil && rec != http.ErrAbortHandler {
				log.Error("panic_recovered",
					"request_id", requestid.FromContext(c),
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"panic", fmt.Sprint(rec),
					"stack", string(debug.Stack()),
				)
				httpx.Fail(c, panicAppError(appEnv, rec))
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

func panicAppError(appEnv string, rec any) *httpx.AppError {
	err := httpx.ErrInternal(fmt.Errorf("%v", rec))
	if appEnv != "production" {
		return err
	}
	return &httpx.AppError{Status: err.Status, Code: err.Code, Message: "خطای داخلی سرور. لطفاً بعداً تلاش کنید."}
}
