package apperr

import (
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

const (
	CodeValidation       = "VALIDATION_ERROR"
	CodeNotFound         = "NOT_FOUND"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeForbidden        = "FORBIDDEN"
	CodeConflict         = "CONFLICT"
	CodeDatabase         = "DATABASE_ERROR"
	CodeInternal         = "INTERNAL_ERROR"
	CodeRouteNotFound    = "ROUTE_NOT_FOUND"
	CodeMethodNotAllowed = "METHOD_NOT_ALLOWED"
)

type AppError struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Cause   error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return e.Cause.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func New(status int, code, message string) *AppError {
	return &AppError{Status: status, Code: code, Message: message}
}

func Wrap(cause error, status int, code, message string) *AppError {
	return &AppError{Status: status, Code: code, Message: message, Cause: cause}
}

func Validation(format string, args ...any) *AppError {
	return New(400, CodeValidation, fmt.Sprintf(format, args...))
}

func NotFound(message string) *AppError {
	return New(404, CodeNotFound, message)
}

func RouteNotFound() *AppError {
	return New(404, CodeRouteNotFound, "مسیر درخواستی یافت نشد.")
}

func MethodNotAllowed() *AppError {
	return New(405, CodeMethodNotAllowed, "متد درخواستی برای این مسیر مجاز نیست.")
}

func Unauthorized(message string) *AppError {
	if message == "" {
		message = "برای انجام این عملیات باید وارد شوید."
	}
	return New(401, CodeUnauthorized, message)
}

func Forbidden(message string) *AppError {
	if message == "" {
		message = "اجازه انجام این عملیات را ندارید."
	}
	return New(403, CodeForbidden, message)
}

func Conflict(format string, args ...any) *AppError {
	return New(409, CodeConflict, fmt.Sprintf(format, args...))
}

func Database(cause error) *AppError {
	return Wrap(cause, 500, CodeDatabase, "خطا در ذخیره‌سازی داده‌ها رخ داد.")
}

func Internal(cause error) *AppError {
	return Wrap(cause, 500, CodeInternal, "خطای غیرمنتظره‌ای رخ داد.")
}

func Normalize(err error) *AppError {
	if err == nil {
		return nil
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
		return NotFound("رکورد مورد نظر یافت نشد.")
	}
	return Internal(err)
}
