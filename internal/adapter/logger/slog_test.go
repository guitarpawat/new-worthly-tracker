package logger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/guitarpawat/worthly-tracker/config"
)

func TestNewCreatesLogFile(t *testing.T) {
	t.Parallel()

	logPath := filepath.Join(t.TempDir(), "logs", "worthly-tracker.log")

	logger, closer, err := New(config.Log{
		Path:  logPath,
		Level: "INFO",
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = closer.Close()
	})

	logger.Info("test log entry", "component", "logger_test")

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if len(content) == 0 {
		t.Fatal("expected log file to contain log output")
	}
}

func TestNewWithoutLogPathDoesNotCreateFile(t *testing.T) {
	tempDir := t.TempDir()
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(currentDir)
	})

	logger, closer, err := New(config.Log{
		Path:  "",
		Level: "INFO",
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = closer.Close()
	})

	logger.Info("stdout only", "component", "logger_test")

	if _, err := os.Stat(filepath.Join(tempDir, "logs", "worthly-tracker.log")); !os.IsNotExist(err) {
		t.Fatalf("expected no log file to be created, got err=%v", err)
	}
}
