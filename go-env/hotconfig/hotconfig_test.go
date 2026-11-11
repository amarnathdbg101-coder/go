package hotconfig_test

import (
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amarnathdbg101-coder/go/go-env/hotconfig"
)

type AppConfig struct {
	Port        int    `json:"port"`
	Environment string `json:"environment"`
	MaxClients  int    `json:"max_clients"`
}

func TestInitialLoadAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	initialJSON := `{"port": 8080, "environment": "production", "max_clients": 100}`
	if err := os.WriteFile(cfgPath, []byte(initialJSON), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := hotconfig.Watch[AppConfig](cfgPath, hotconfig.WithPollInterval[AppConfig](50*time.Millisecond))
	if err != nil {
		t.Fatalf("failed to watch config: %v", err)
	}
	defer cfg.Close()

	current := cfg.Get()
	if current.Port != 8080 || current.Environment != "production" {
		t.Fatalf("unexpected config values: %+v", current)
	}
}

func TestDynamicHotReload(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	v1 := `{"port": 8080, "environment": "staging", "max_clients": 50}`
	_ = os.WriteFile(cfgPath, []byte(v1), 0644)

	var changeCount int32
	cfg, err := hotconfig.Watch[AppConfig](
		cfgPath,
		hotconfig.WithPollInterval[AppConfig](10*time.Millisecond),
		hotconfig.WithOnChange[AppConfig](func(oldCfg, newCfg *AppConfig) {
			atomic.AddInt32(&changeCount, 1)
		}),
	)
	if err != nil {
		t.Fatalf("watch failed: %v", err)
	}
	defer cfg.Close()

	if cfg.Get().Port != 8080 {
		t.Fatalf("expected port 8080, got %d", cfg.Get().Port)
	}

	// Update file on disk
	time.Sleep(20 * time.Millisecond)
	v2 := `{"port": 9090, "environment": "production", "max_clients": 200}`
	_ = os.WriteFile(cfgPath, []byte(v2), 0644)

	// Wait for background reload
	time.Sleep(60 * time.Millisecond)

	updated := cfg.Get()
	if updated.Port != 9090 || updated.Environment != "production" {
		t.Fatalf("expected port 9090 after hot reload, got %+v", updated)
	}
	if atomic.LoadInt32(&changeCount) != 1 {
		t.Fatalf("expected 1 change callback, got %d", atomic.LoadInt32(&changeCount))
	}
}

func TestCorruptFilePreservesOldConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	v1 := `{"port": 8080, "environment": "production", "max_clients": 50}`
	_ = os.WriteFile(cfgPath, []byte(v1), 0644)

	cfg, _ := hotconfig.Watch[AppConfig](cfgPath, hotconfig.WithPollInterval[AppConfig](10*time.Millisecond))
	defer cfg.Close()

	// Write garbage / half-written truncated JSON
	time.Sleep(20 * time.Millisecond)
	_ = os.WriteFile(cfgPath, []byte(`{"port": 9999, "corrupt...`), 0644)

	time.Sleep(50 * time.Millisecond)

	// Old configuration MUST be preserved safely
	if cfg.Get().Port != 8080 {
		t.Fatalf("expected old config 8080 to be preserved after corruption, got %d", cfg.Get().Port)
	}
}

func TestValidatorRejection(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	v1 := `{"port": 8080, "environment": "production", "max_clients": 50}`
	_ = os.WriteFile(cfgPath, []byte(v1), 0644)

	cfg, _ := hotconfig.Watch[AppConfig](
		cfgPath,
		hotconfig.WithPollInterval[AppConfig](10*time.Millisecond),
		hotconfig.WithValidator[AppConfig](func(c *AppConfig) error {
			if c.Port < 1024 || c.Port > 65535 {
				return errors.New("port out of range")
			}
			return nil
		}),
	)
	defer cfg.Close()

	// Attempt to set invalid port (80)
	time.Sleep(20 * time.Millisecond)
	_ = os.WriteFile(cfgPath, []byte(`{"port": 80, "environment": "production", "max_clients": 50}`), 0644)

	time.Sleep(50 * time.Millisecond)

	// Must be rejected by validator and preserve 8080
	if cfg.Get().Port != 8080 {
		t.Fatalf("validator failed to reject invalid port, got %d", cfg.Get().Port)
	}
}

