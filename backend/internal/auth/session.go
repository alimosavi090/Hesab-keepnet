package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"github.com/ali/hesab-keepnet/backend/internal/models"
	"github.com/ali/hesab-keepnet/backend/internal/passwordhash"
	"gorm.io/gorm"
)

const (
	SessionCookie = "hesab_session"
	CSRFCookie    = "hesab_csrf"
	CSRFHeader    = "X-CSRF-Token"

	defaultTTL = 7 * 24 * time.Hour
)

type Manager struct {
	db  *gorm.DB
	ttl time.Duration
}

func NewManager(db *gorm.DB) *Manager {
	return &Manager{db: db, ttl: defaultTTL}
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func newToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", errors.New("generate token")
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func NewToken() string {
	token, err := newToken()
	if err != nil {
		panic(err)
	}
	return token
}

type SessionInfo struct {
	Token     string
	ExpiresAt time.Time
	User      *models.User
}

func (m *Manager) Login(ctx context.Context, username, password, ip, userAgent string) (*SessionInfo, error) {
	var user models.User
	err := m.db.WithContext(ctx).
		Where("username = ? AND deleted_at IS NULL", username).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.Unauthorized("نام کاربری یا گذرواژه نادرست است.")
	}
	if err != nil {
		return nil, apperr.Database(err)
	}
	if !user.IsActive {
		return nil, apperr.Forbidden("حساب کاربری غیرفعال است.")
	}

	ok, err := passwordhash.Verify(password, user.PasswordHash)
	if err != nil || !ok {
		return nil, apperr.Unauthorized("نام کاربری یا گذرواژه نادرست است.")
	}

	token, expiresAt, err := m.createSession(ctx, user.ID, ip, userAgent)
	if err != nil {
		return nil, err
	}
	return &SessionInfo{Token: token, ExpiresAt: expiresAt, User: &user}, nil
}

func (m *Manager) createSession(ctx context.Context, userID int64, ip, userAgent string) (string, time.Time, error) {
	token, err := newToken()
	if err != nil {
		return "", time.Time{}, apperr.Internal(err)
	}
	expiresAt := time.Now().UTC().Add(m.ttl)

	session := models.Session{
		TokenHash: HashToken(token),
		UserID:    userID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
	}
	if ip != "" {
		session.IPAddress = ip
	}
	if userAgent != "" {
		session.UserAgent = userAgent
	}

	if err := m.db.WithContext(ctx).Create(&session).Error; err != nil {
		return "", time.Time{}, apperr.Database(err)
	}
	return token, expiresAt, nil
}

func (m *Manager) Validate(ctx context.Context, token string) (*models.User, error) {
	hashed := HashToken(token)

	var session models.Session
	err := m.db.WithContext(ctx).
		Where("token_hash = ? AND expires_at > ?", hashed, time.Now().UTC()).
		First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.Unauthorized("")
	}
	if err != nil {
		return nil, apperr.Database(err)
	}

	var user models.User
	err = m.db.WithContext(ctx).
		Where("id = ? AND is_active = ? AND deleted_at IS NULL", session.UserID, true).
		First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.Unauthorized("")
	}
	if err != nil {
		return nil, apperr.Database(err)
	}

	if remaining := time.Until(session.ExpiresAt); remaining < m.ttl/2 {
		_ = m.db.WithContext(ctx).Model(&models.Session{}).
			Where("token_hash = ?", hashed).
			Update("expires_at", time.Now().UTC().Add(m.ttl)).Error
	}

	return &user, nil
}

func (m *Manager) Destroy(ctx context.Context, token string) error {
	if err := m.db.WithContext(ctx).
		Where("token_hash = ?", HashToken(token)).
		Delete(&models.Session{}).Error; err != nil {
		return apperr.Database(err)
	}
	return nil
}

func (m *Manager) DeleteExpired(ctx context.Context) (int64, error) {
	result := m.db.WithContext(ctx).
		Where("expires_at < ?", time.Now().UTC()).
		Delete(&models.Session{})
	return result.RowsAffected, result.Error
}
