package handlers

import (
	"net/http"
	"time"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/auth"
	"github.com/ali/hesab-keepnet/backend/internal/httpx"
	"github.com/ali/hesab-keepnet/backend/internal/models"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	manager      *auth.Manager
	cookieSecure bool
}

func NewAuthHandler(manager *auth.Manager, cookieSecure bool) *AuthHandler {
	return &AuthHandler{manager: manager, cookieSecure: cookieSecure}
}

type publicUser struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.HandleError(c, apperr.Validation("نام کاربری و گذرواژه الزامی است."))
		return
	}

	info, err := h.manager.Login(
		c.Request.Context(),
		req.Username, req.Password,
		c.ClientIP(), c.Request.UserAgent(),
	)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	h.setSessionCookie(c, info.Token, info.ExpiresAt)
	h.setCSRFCookie(c, auth.NewToken())

	httpx.OK(c, http.StatusOK, publicUser{
		ID:          info.User.ID,
		Username:    info.User.Username,
		DisplayName: info.User.DisplayName,
		Role:        info.User.Role,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	if token, err := c.Cookie(auth.SessionCookie); err == nil && token != "" {
		if err := h.manager.Destroy(c.Request.Context(), token); err != nil {
			httpx.HandleError(c, err)
			return
		}
	}
	h.clearCookies(c)
	httpx.OK(c, http.StatusOK, gin.H{"logged_out": true})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userAny, _ := c.Get("current_user")
	user, ok := userAny.(*models.User)
	if !ok {
		httpx.HandleError(c, apperr.Unauthorized(""))
		return
	}
	httpx.OK(c, http.StatusOK, publicUser{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Role:        user.Role,
	})
}

func (h *AuthHandler) CSRF(c *gin.Context) {
	existing, err := c.Cookie(auth.CSRFCookie)
	if err != nil || existing == "" {
		existing = auth.NewToken()
		h.setCSRFCookie(c, existing)
	}
	httpx.OK(c, http.StatusOK, gin.H{"token": existing})
}

func (h *AuthHandler) setSessionCookie(c *gin.Context, token string, expires time.Time) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     auth.SessionCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  expires,
	})
}

func (h *AuthHandler) setCSRFCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     auth.CSRFCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: false,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})
}

func (h *AuthHandler) clearCookies(c *gin.Context) {
	expired := time.Unix(0, 0)
	for _, name := range []string{auth.SessionCookie, auth.CSRFCookie} {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			HttpOnly: name == auth.SessionCookie,
			Secure:   h.cookieSecure,
			SameSite: http.SameSiteLaxMode,
			Expires:  expired,
			MaxAge:   -1,
		})
	}
}
