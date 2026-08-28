package middleware

import (
	"net/http"
	"time"

	"github.com/ali/hesab-keepnet/backend/internal/requestid"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const (
	corsMaxAge = 12 * time.Hour
	csrfHeader = "X-CSRF-Token"
)

func CORS(origins []string) gin.HandlerFunc {
	config := cors.Config{
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", requestid.Header, csrfHeader},
		ExposeHeaders:    []string{requestid.Header},
		AllowCredentials: true,
		MaxAge:           corsMaxAge,
	}
	if len(origins) == 0 {
		config.AllowOriginFunc = func(string) bool { return false }
	} else if len(origins) == 1 && origins[0] == "*" {
		config.AllowOrigins = origins
		config.AllowCredentials = false
	} else {
		config.AllowOrigins = origins
	}
	return cors.New(config)
}
