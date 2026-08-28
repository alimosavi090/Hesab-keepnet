package middleware

import (
	"github.com/ali/hesab-keepnet/backend/internal/apperr"
)

func apperrUnauthorized() *apperr.AppError {
	return apperr.Unauthorized("برای انجام این عملیات باید وارد شوید.")
}

func apperrTooManyRequests() *apperr.AppError {
	return apperr.New(429, "RATE_LIMITED", "تعداد درخواست‌ها بیش از حد مجاز است. کمی بعد تلاش کنید.")
}
