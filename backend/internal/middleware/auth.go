package middleware

import (
	"crypto/subtle"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/auth"
	"github.com/ali/hesab-keepnet/backend/internal/ctxkeys"
	"github.com/gin-gonic/gin"
)

func Authenticate(manager *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(auth.SessionCookie)
		if err != nil || token == "" {
			c.AbortWithStatusJSON(401, unauthorizedEnvelope())
			return
		}

		user, err := manager.Validate(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(401, unauthorizedEnvelope())
			return
		}

		ctx := ctxkeys.WithUserID(c.Request.Context(), user.ID)
		c.Request = c.Request.WithContext(ctx)
		c.Set("current_user", user)
		c.Next()
	}
}

func RequireCSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case "GET", "HEAD", "OPTIONS":
			c.Next()
			return
		}

		headerToken := c.GetHeader(auth.CSRFHeader)
		cookieToken, _ := c.Cookie(auth.CSRFCookie)
		if headerToken == "" || cookieToken == "" ||
			subtle.ConstantTimeCompare([]byte(headerToken), []byte(cookieToken)) != 1 {
			c.AbortWithStatusJSON(403, envelope(apperr.Forbidden("توکن CSRF نامعتبر است.")))
			return
		}
		c.Next()
	}
}

func CurrentUserID(c *gin.Context) (int64, bool) {
	return ctxkeys.UserIDFrom(c.Request.Context())
}
