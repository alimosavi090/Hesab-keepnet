package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	gormsqlite "github.com/glebarez/sqlite"
)

const backupPrefix = "accounting-backup-"

func main() {
	dbPath := flag.String("db", "./data/accounting.db", "path to source sqlite database")
	outDir := flag.String("out", "./backups", "directory for backup files")
	keep := flag.Int("keep", 14, "number of most recent backups to retain")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if err := run(*dbPath, *outDir, *keep, logger); err != nil {
		logger.Error("backup failed", "err", err)
		os.Exit(1)
	}
}

func run(dbPath, outDir string, keep int, logger *slog.Logger) error {
	if keep < 1 {
		return fmt.Errorf("--keep must be at least 1")
	}
	if _, err := os.Stat(dbPath); err != nil {
		return fmt.Errorf("source database not accessible: %w", err)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	sourceDSN := fmt.Sprintf("file:%s?mode=ro&_pragma=busy_timeout(5000)", dbPath)
	source, err := sql.Open(gormsqlite.DriverName, sourceDSN)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer source.Close()
	if err := source.Ping(); err != nil {
		return fmt.Errorf("ping source: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	targetPath := filepath.Join(outDir, backupPrefix+timestamp+".db")

	if _, err := source.Exec(fmt.Sprintf("VACUUM INTO '%s'", strings.ReplaceAll(targetPath, "'", "''"))); err != nil {
		return fmt.Errorf("vacuum into %s: %w", targetPath, err)
	}

	if err := verifyIntegrity(targetPath); err != nil {
		return fmt.Errorf("integrity check failed for %s: %w", targetPath, err)
	}
	info, err := os.Stat(targetPath)
	if err != nil {
		return fmt.Errorf("stat backup: %w", err)
	}
	logger.Info("backup created", "path", targetPath, "size_bytes", info.Size())

	pruned, err := pruneOldBackups(outDir, keep)
	if err != nil {
		logger.Warn("prune failed", "err", err)
	} else if pruned > 0 {
		logger.Info("old backups pruned", "count", pruned)
	}

	return nil
}

func verifyIntegrity(path string) error {
	checkDB, err := sql.Open(gormsqlite.DriverName, path)
	if err != nil {
		return err
	}
	defer checkDB.Close()

	var result string
	if err := checkDB.QueryRow("PRAGMA integrity_check").Scan(&result); err != nil {
		return err
	}
	if result != "ok" {
		return fmt.Errorf("integrity_check returned %q", result)
	}
	return nil
}

func pruneOldBackups(outDir string, keep int) (int, error) {
	entries, err := os.ReadDir(outDir)
	if err != nil {
		return 0, err
	}

	var backups []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), backupPrefix) && strings.HasSuffix(entry.Name(), ".db") {
			backups = append(backups, entry.Name())
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(backups)))

	if len(backups) <= keep {
		return 0, nil
	}

	pruned := 0
	for _, name := range backups[keep:] {
		if err := os.Remove(filepath.Join(outDir, name)); err != nil {
			return pruned, err
		}
		pruned++
	}
	return pruned, nil
}
