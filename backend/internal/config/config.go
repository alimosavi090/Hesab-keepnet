package config

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	EnvDevelopment = "development"
	EnvTest        = "test"
	EnvProduction  = "production"

	minSecretLength = 16
)

type Config struct {
	AppEnv                 string
	AppPort                int
	DatabasePath           string
	BackupDir              string
	BackupIntervalHours    int
	CorsOrigins            []string
	SessionSecret          string
	SessionSecretEphemeral bool
	CookieSecure           bool
	AdminUsername          string
	AdminPassword          string
}

func Load() (*Config, error) {
	_ = loadDotenv(".env")

	cfg := &Config{
		AppEnv:       getEnv("APP_ENV", EnvDevelopment),
		DatabasePath: getEnv("DATABASE_PATH", "./data/accounting.db"),
		CorsOrigins:  parseOrigins(getEnv("CORS_ORIGIN", "http://localhost:3000")),
		CookieSecure: getEnvBool("COOKIE_SECURE", false),
	}

	cfg.BackupDir = getEnv("BACKUP_DIR", "")
	if cfg.BackupDir == "" {
		cfg.BackupDir = filepath.Join(filepath.Dir(cfg.DatabasePath), "backups")
	}
	if hours, err := strconv.Atoi(getEnv("BACKUP_INTERVAL_HOURS", "24")); err == nil && hours > 0 {
		cfg.BackupIntervalHours = hours
	} else {
		cfg.BackupIntervalHours = 24
	}

	if err := cfg.parsePort(); err != nil {
		return nil, err
	}

	if err := cfg.validateEnv(); err != nil {
		return nil, err
	}
	if err := cfg.resolveSessionSecret(); err != nil {
		return nil, err
	}
	if cfg.AppEnv == EnvProduction && os.Getenv("COOKIE_SECURE") == "" {
		cfg.CookieSecure = true
	}

	// Admin seeding is allowed in every environment when explicitly provided;
	// the seed itself only runs while the users table is empty, so this is
	// safe in production (e.g. after restoring a database without users).
	cfg.AdminUsername = os.Getenv("ADMIN_USERNAME")
	cfg.AdminPassword = os.Getenv("ADMIN_PASSWORD")

	return cfg, nil
}

func (c *Config) IsProd() bool {
	return c.AppEnv == EnvProduction
}

func (c *Config) Addr() string {
	return fmt.Sprintf(":%d", c.AppPort)
}

func (c *Config) parsePort() error {
	raw := os.Getenv("APP_PORT")
	if raw == "" {
		c.AppPort = 8080
		return nil
	}
	port, err := strconv.Atoi(raw)
	if err != nil {
		return fmt.Errorf("invalid APP_PORT %q: must be a number between 1 and 65535", raw)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid APP_PORT %d: must be between 1 and 65535", port)
	}
	c.AppPort = port
	return nil
}

func (c *Config) validateEnv() error {
	switch c.AppEnv {
	case EnvDevelopment, EnvTest, EnvProduction:
		return nil
	default:
		return fmt.Errorf("invalid APP_ENV %q: must be one of development, test, production", c.AppEnv)
	}
}

func (c *Config) resolveSessionSecret() error {
	switch {
	case os.Getenv("SESSION_SECRET") != "":
		c.SessionSecret = os.Getenv("SESSION_SECRET")
		if len(c.SessionSecret) < minSecretLength {
			return errors.New("SESSION_SECRET is too short: must be at least 16 characters")
		}
	case c.AppEnv == EnvProduction:
		return errors.New("SESSION_SECRET is required when APP_ENV=production")
	default:
		buf := make([]byte, 32)
		if _, err := rand.Read(buf); err != nil {
			return fmt.Errorf("generate ephemeral session secret: %w", err)
		}
		c.SessionSecret = hex.EncodeToString(buf)
		c.SessionSecretEphemeral = true
	}
	return nil
}

func parseOrigins(raw string) []string {
	var origins []string
	for _, part := range strings.Split(raw, ",") {
		if part = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(part), "/")); part != "" {
			origins = append(origins, part)
		}
	}
	return origins
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v := strings.ToLower(os.Getenv(key))
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
