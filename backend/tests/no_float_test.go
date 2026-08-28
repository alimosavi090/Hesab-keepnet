package tests

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var floatPattern = regexp.MustCompile(`\bfloat(32|64)\b`)

func TestNoFloatingPointInMoneyPackages(t *testing.T) {
	moneyDirs := []string{
		filepath.Join("..", "internal", "models"),
		filepath.Join("..", "internal", "enums"),
		filepath.Join("..", "internal", "repository"),
		filepath.Join("..", "internal", "services"),
	}

	for _, dir := range moneyDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
				continue
			}
			path := filepath.Join(dir, entry.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if match := floatPattern.Find(content); match != nil {
				t.Errorf("%s uses floating point (%q); money must be integer-only", path, string(match))
			}
		}
	}
}
