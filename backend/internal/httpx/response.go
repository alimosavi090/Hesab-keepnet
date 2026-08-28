package httpx

import (
	"fmt"
	"log/slog"

	"github.com/ali/hesab-keepnet/backend/internal/requestid"
	"github.com/gin-gonic/gin"
)

var productionMode bool

func SetProduction(prod bool) {
	productionMode = prod
}

type Response struct {
	Success bool      `json:"success"`
	Data    any       `json:"data"`
	Error   *AppError `json:"error"`
}

func OK(c *gin.Context, status int, data any) {
	c.JSON(status, Response{Success: true, Data: data})
}

func Fail(c *gin.Context, appErr *AppError) {
	c.JSON(appErr.Status, Response{Success: false, Data: nil, Error: sanitize(appErr)})
}

func HandleError(c *gin.Context, err error) {
	appErr := Normalize(err)
	if appErr.Status >= 500 && !productionMode && appErr.Cause != nil {
		appErr.Message = fmt.Sprintf("%s (%s)", appErr.Message, appErr.Error())
	}

	slog.ErrorContext(c.Request.Context(), "api_error",
		"code", appErr.Code,
		"status", appErr.Status,
		"request_id", requestid.FromContext(c),
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"err", appErr.Error(),
	)

	Fail(c, appErr)
}

func sanitize(appErr *AppError) *AppError {
	if !productionMode {
		return appErr
	}
	switch appErr.Code {
	case CodeInternal, CodeDatabase:
		return &AppError{Status: appErr.Status, Code: appErr.Code, Message: "خطای داخلی سرور. لطفاً بعداً تلاش کنید."}
	default:
		return appErr
	}
}
