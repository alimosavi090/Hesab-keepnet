package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ali/hesab-keepnet/backend/internal/apperr"
	"gorm.io/gorm"
)

type BackupFile struct {
	Name      string    `json:"name"`
	SizeBytes int64     `json:"size_bytes"`
	CreatedAt time.Time `json:"created_at"`
	IsAuto    bool      `json:"is_auto"`
}

const (
	autoBackupPrefix     = "auto-"
	manualBackupPrefix   = "backup-"
	backupFileExt        = ".db"
	defaultBackupRetain  = 20
)

// BackupService produces consistent SQLite snapshots via `VACUUM INTO` and
// manages retention on disk. Backups never live inside the database itself.
type BackupService struct {
	db    *gorm.DB
	dir   string
	retain int
}

func NewBackupService(db *gorm.DB, dir string) *BackupService {
	return &BackupService{db: db, dir: dir, retain: defaultBackupRetain}
}

func (s *BackupService) validName(name string) bool {
	if strings.ContainsAny(name, "/\\") || strings.Contains(name, "..") {
		return false
	}
	base := filepath.Base(name)
	return (strings.HasPrefix(base, autoBackupPrefix) || strings.HasPrefix(base, manualBackupPrefix)) &&
		strings.HasSuffix(base, backupFileExt)
}

// Create writes a fresh snapshot. `auto` controls the filename prefix.
func (s *BackupService) Create(ctx context.Context, auto bool) (*BackupFile, error) {
	if err := os.MkdirAll(s.dir, 0o750); err != nil {
		return nil, apperr.Internal(fmt.Errorf("ایجاد پوشه پشتیبان ناموفق بود: %w", err))
	}

	prefix := manualBackupPrefix
	if auto {
		prefix = autoBackupPrefix
	}
	name := fmt.Sprintf("%s%s%s", prefix, time.Now().UTC().Format("20060102_150405"), backupFileExt)
	target := filepath.Join(s.dir, name)

	if _, err := os.Stat(target); err == nil {
		return nil, apperr.Conflict("پشتیبانی هم‌نام وجود دارد؛ چند لحظه بعد تلاش کنید.")
	}

	if err := s.db.WithContext(ctx).Exec("VACUUM INTO ?", target).Error; err != nil {
		return nil, apperr.Database(fmt.Errorf("VACUUM INTO: %w", err))
	}

	if err := s.prune(); err != nil {
		// Retention failure must not fail the backup creation.
		os.Remove(target)
		return nil, err
	}

	info, err := os.Stat(target)
	if err != nil {
		return nil, apperr.Database(err)
	}
	return &BackupFile{Name: name, SizeBytes: info.Size(), CreatedAt: info.ModTime().UTC(), IsAuto: auto}, nil
}

func (s *BackupService) prune() error {
	files, err := s.list()
	if err != nil {
		return err
	}
	if len(files) <= s.retain {
		return nil
	}
	sort.Slice(files, func(i, j int) bool { return files[i].CreatedAt.Before(files[j].CreatedAt) })
	for _, f := range files[:len(files)-s.retain] {
		_ = os.Remove(filepath.Join(s.dir, f.Name))
	}
	return nil
}

// LastAuto returns the newest automatic backup timestamp (zero if none).
func (s *BackupService) LastAuto(ctx context.Context) (time.Time, error) {
	files, err := s.list()
	if err != nil {
		return time.Time{}, err
	}
	var last time.Time
	for _, f := range files {
		if f.IsAuto && f.CreatedAt.After(last) {
			last = f.CreatedAt
		}
	}
	return last, nil
}

// AutoDue reports whether a scheduled backup is needed.
func (s *BackupService) AutoDue(interval time.Duration) bool {
	last, err := s.LastAuto(context.Background())
	if err != nil || last.IsZero() {
		return true
	}
	return time.Since(last) >= interval
}

func (s *BackupService) List() ([]BackupFile, error) {
	return s.list()
}

func (s *BackupService) list() ([]BackupFile, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []BackupFile{}, nil
		}
		return nil, apperr.Database(err)
	}
	files := make([]BackupFile, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !s.validName(e.Name()) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, BackupFile{
			Name:      e.Name(),
			SizeBytes: info.Size(),
			CreatedAt: info.ModTime().UTC(),
			IsAuto:    strings.HasPrefix(e.Name(), autoBackupPrefix),
		})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].CreatedAt.After(files[j].CreatedAt) })
	return files, nil
}

// Path resolves a safe absolute path for download/delete; rejects traversal.
func (s *BackupService) Path(name string) (string, error) {
	if !s.validName(name) {
		return "", apperr.Validation("نام فایل پشتیبان نامعتبر است.")
	}
	path := filepath.Join(s.dir, filepath.Base(name))
	if _, err := os.Stat(path); err != nil {
		return "", apperr.NotFound("فایل پشتیبان یافت نشد.")
	}
	return path, nil
}

func (s *BackupService) Delete(name string) error {
	path, err := s.Path(name)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return apperr.Database(err)
	}
	return nil
}
