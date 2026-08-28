package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ali/hesab-keepnet/backend/internal/config"
	"github.com/ali/hesab-keepnet/backend/internal/database"
	"github.com/ali/hesab-keepnet/backend/internal/routes"
	"github.com/ali/hesab-keepnet/backend/internal/services"
	"github.com/ali/hesab-keepnet/backend/internal/version"
	"github.com/ali/hesab-keepnet/backend/migrations"
)

const shutdownTimeout = 10 * time.Second

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := newLogger(cfg)
	slog.SetDefault(log)

	if cfg.SessionSecretEphemeral {
		log.Warn("SESSION_SECRET is not set; generated an ephemeral secret; sessions will not survive restarts", "app_env", cfg.AppEnv)
	}

	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Error("close database failed", "err", closeErr)
		} else {
			log.Info("database closed")
		}
	}()

	applied, err := database.Migrate(migrations.FS, db)
	if err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	if len(applied) > 0 {
		log.Info("database ready", "applied_migrations", len(applied), "latest_version", applied[len(applied)-1].Version)
	}

	if err := database.Seed(context.Background(), db.DB, cfg.AdminUsername, cfg.AdminPassword); err != nil {
		return fmt.Errorf("seed database: %w", err)
	}
	if cfg.AdminUsername != "" {
		log.Info("admin user seeded", "username", cfg.AdminUsername)
	}

	router := routes.NewRouter(cfg, db, log)

	// Automatic database backups: run one at startup if due, then check hourly.
	backupSvc := services.NewBackupService(db.DB, cfg.BackupDir)
	interval := time.Duration(cfg.BackupIntervalHours) * time.Hour
	go runAutoBackups(log, backupSvc, interval)

	server := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       20 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.ListenAndServe()
	}()

	log.Info("server started",
		"addr", cfg.Addr(),
		"app_env", cfg.AppEnv,
		"database_path", cfg.DatabasePath,
		"version", version.Version,
	)

	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server: %w", err)
		}
	case <-ctx.Done():
		log.Info("shutdown signal received; draining connections")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed; forcing exit", "err", err)
	} else {
		log.Info("server stopped gracefully")
	}

	return nil
}

func newLogger(cfg *config.Config) *slog.Logger {
	level := slog.LevelInfo
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(handler).With("service", "hesab-backend", "env", cfg.AppEnv)
}

func runAutoBackups(log *slog.Logger, backups *services.BackupService, interval time.Duration) {
	// Catch-up: if the last automatic snapshot is older than the interval,
	// take one immediately after boot.
	if backups.AutoDue(interval) {
		if _, err := backups.Create(context.Background(), true); err != nil {
			log.Error("startup auto-backup failed", "err", err)
		} else {
			log.Info("startup auto-backup created", "dir", "configured")
		}
	}
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		if !backups.AutoDue(interval) {
			continue
		}
		if _, err := backups.Create(context.Background(), true); err != nil {
			log.Error("scheduled auto-backup failed", "err", err)
		} else {
			log.Info("scheduled auto-backup created")
		}
	}
}
