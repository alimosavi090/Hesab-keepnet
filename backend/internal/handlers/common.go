package handlers

import (
	"net/http"
	"strconv"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/httpx"
	"github.com/ali/hesab-keepnet/backend/internal/services"
	"github.com/gin-gonic/gin"
)

type pagedBody[T any] struct {
	Items []T           `json:"items"`
	Meta  services.Meta `json:"meta"`
}

func okPaged[T any](c *gin.Context, items []T, total int64, page services.PageQuery) {
	_, pageSize := page.Normalized()
	if page.Page < 1 {
		page.Page = 1
	}
	if page.PageSize < 1 {
		page.PageSize = services.DefaultPageSize
	}
	if page.PageSize > services.MaxPageSize {
		page.PageSize = services.MaxPageSize
	}
	httpx.OK(c, http.StatusOK, pagedBody[T]{
		Items: items,
		Meta:  services.Meta{Total: total, Page: page.Page, PageSize: pageSize},
	})
}

func bindJSON[T any](c *gin.Context) (*T, error) {
	var body T
	if err := c.ShouldBindJSON(&body); err != nil {
		return nil, apperr.Validation("ورودی نامعتبر است: %s", firstError(err))
	}
	return &body, nil
}

func queryPage(c *gin.Context) (services.PageQuery, error) {
	var pq services.PageQuery
	pq.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	pq.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))
	return pq, nil
}

func firstError(err error) string {
	msg := err.Error()
	if len(msg) > 120 {
		msg = msg[:120]
	}
	return msg
}
