package httpx

import "github.com/ali/hesab-keepnet/backend/internal/apperr"

type AppError = apperr.AppError

const (
	CodeValidation       = apperr.CodeValidation
	CodeNotFound         = apperr.CodeNotFound
	CodeUnauthorized     = apperr.CodeUnauthorized
	CodeForbidden        = apperr.CodeForbidden
	CodeConflict         = apperr.CodeConflict
	CodeDatabase         = apperr.CodeDatabase
	CodeInternal         = apperr.CodeInternal
	CodeRouteNotFound    = apperr.CodeRouteNotFound
	CodeMethodNotAllowed = apperr.CodeMethodNotAllowed
)

var (
	ErrValidation       = apperr.Validation
	ErrNotFound         = apperr.NotFound
	ErrRouteNotFound    = apperr.RouteNotFound
	ErrMethodNotAllowed = apperr.MethodNotAllowed
	ErrUnauthorized     = apperr.Unauthorized
	ErrForbidden        = apperr.Forbidden
	ErrConflict         = apperr.Conflict
	ErrDatabase         = apperr.Database
	ErrInternal         = apperr.Internal

	Normalize = apperr.Normalize
)
