package requestid

import (
	"context"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	Header     = "X-Request-ID"
	ContextKey = "request_id"
)

var allowedPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{8,64}$`)

func FromContext(c *gin.Context) string {
	if v, ok := c.Get(ContextKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func FromCtx(ctx context.Context) string {
	if v, ok := ctx.Value(ContextKey).(string); ok {
		return v
	}
	return ""
}

func New() string {
	return uuid.NewString()
}

func IsValid(id string) bool {
	return allowedPattern.MatchString(id)
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(Header)
		if !IsValid(id) {
			id = New()
		}
		c.Set(ContextKey, id)
		c.Writer.Header().Set(Header, id)
		c.Next()
	}
}
