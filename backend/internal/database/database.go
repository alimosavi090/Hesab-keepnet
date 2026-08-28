package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const (
	busyTimeoutMs = 5000
	maxOpenConns  = 10
	maxIdleConns  = 5
	connIdleTTL   = time.Hour
	pingTimeout   = 2 * time.Second
)

type DB struct {
	DB   *gorm.DB
	Path string
}

func Open(path string) (*DB, error) {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}

	dsn := fmt.Sprintf(
		"file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(%d)",
		path, busyTimeoutMs,
	)

	gdb, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open sqlite database %q: %w", path, err)
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, fmt.Errorf("unwrap sql db: %w", err)
	}
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxIdleTime(connIdleTTL)

	return &DB{DB: gdb, Path: path}, nil
}

func (d *DB) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (d *DB) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (d *DB) PragmaText(name string) (string, error) {
	var value string
	if err := d.DB.Raw("PRAGMA " + name).Scan(&value).Error; err != nil {
		return "", err
	}
	return value, nil
}

func (d *DB) PragmaInt(name string) (int64, error) {
	var value int64
	if err := d.DB.Raw("PRAGMA " + name).Scan(&value).Error; err != nil {
		return 0, err
	}
	return value, nil
}
