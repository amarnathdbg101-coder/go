package envconfig

import (
	"errors"
	"testing"
	"time"
)

type DatabaseConfig struct {
	Host     string `env:"HOST" default:"localhost"`
	Port     int    `env:"PORT" default:"5432"`
	Password string `env:"PASSWORD" required:"true" mask:"true"`
}

type AppConfig struct {
	AppName     string         `env:"NAME" default:"MyApp"`
	Debug       bool           `env:"DEBUG" default:"false"`
	Timeout     time.Duration  `env:"TIMEOUT" default:"30s"`
	RateLimit   float64        `env:"RATE_LIMIT" default:"100.5"`
	AllowedIPs  []string       `env:"ALLOWED_IPS" split:";"`
	WorkerPorts []int          `env:"PORTS" split:","`
	DB          DatabaseConfig `env:"DB"`
}

func mockLookup(envMap map[string]string) LookupFunc {
	return func(key string) (string, bool) {
		val, ok := envMap[key]
		return val, ok
	}
}

func TestProcess_Success(t *testing.T) {
	envs := map[string]string{
		"APP_NAME":        "OrderService",
		"APP_DEBUG":       "true",
		"APP_TIMEOUT":     "15s",
		"APP_RATE_LIMIT":  "250.75",
		"APP_ALLOWED_IPS": "127.0.0.1;10.0.0.1;192.168.1.1",
		"APP_PORTS":       "8001,8002,8003",
		"APP_DB_HOST":     "postgres-master",
		"APP_DB_PORT":     "5433",
		"APP_DB_PASSWORD": "supersecretpassword123",
	}

	var cfg AppConfig
	err := ProcessWithLookup("APP", &cfg, mockLookup(envs))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.AppName != "OrderService" {
		t.Errorf("expected AppName 'OrderService', got '%s'", cfg.AppName)
	}
	if !cfg.Debug {
		t.Errorf("expected Debug true, got false")
	}
	if cfg.Timeout != 15*time.Second {
		t.Errorf("expected Timeout 15s, got %v", cfg.Timeout)
	}
	if cfg.RateLimit != 250.75 {
		t.Errorf("expected RateLimit 250.75, got %f", cfg.RateLimit)
	}
	if len(cfg.AllowedIPs) != 3 || cfg.AllowedIPs[1] != "10.0.0.1" {
		t.Errorf("expected AllowedIPs [127.0.0.1 10.0.0.1 192.168.1.1], got %v", cfg.AllowedIPs)
	}
	if len(cfg.WorkerPorts) != 3 || cfg.WorkerPorts[2] != 8003 {
		t.Errorf("expected WorkerPorts [8001 8002 8003], got %v", cfg.WorkerPorts)
	}
	if cfg.DB.Host != "postgres-master" || cfg.DB.Port != 5433 || cfg.DB.Password != "supersecretpassword123" {
		t.Errorf("nested DB config mismatch: %+v", cfg.DB)
	}
}

func TestProcess_Defaults(t *testing.T) {
	// Only required password is provided
	envs := map[string]string{
		"PASSWORD": "mypassword",
	}

	var cfg DatabaseConfig
	err := ProcessWithLookup("", &cfg, mockLookup(envs))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Host != "localhost" {
		t.Errorf("expected default host 'localhost', got '%s'", cfg.Host)
	}
	if cfg.Port != 5432 {
		t.Errorf("expected default port 5432, got %d", cfg.Port)
	}
	if cfg.Password != "mypassword" {
		t.Errorf("expected password 'mypassword', got '%s'", cfg.Password)
	}
}

func TestProcess_RequiredMissing(t *testing.T) {
	envs := map[string]string{
		"DB_HOST": "localhost",
	}

	var cfg DatabaseConfig
	err := ProcessWithLookup("DB", &cfg, mockLookup(envs))
	if err == nil {
		t.Fatalf("expected error for missing required field, got nil")
	}
	if !errors.Is(err, ErrMissingRequired) {
		t.Errorf("expected ErrMissingRequired, got: %v", err)
	}
}

func TestProcess_NotPointerToStruct(t *testing.T) {
	var cfg AppConfig
	err := Process("", cfg) // passed value instead of pointer
	if !errors.Is(err, ErrNotPointerToStruct) {
		t.Errorf("expected ErrNotPointerToStruct, got: %v", err)
	}

	var num int = 42
	err = Process("", &num)
	if !errors.Is(err, ErrNotPointerToStruct) {
		t.Errorf("expected ErrNotPointerToStruct, got: %v", err)
	}
}

func TestProcess_InvalidParsing(t *testing.T) {
	type BadConfig struct {
		Port int `env:"PORT"`
	}

	envs := map[string]string{"PORT": "not-a-number"}
	var cfg BadConfig
	err := ProcessWithLookup("", &cfg, mockLookup(envs))
	if err == nil {
		t.Fatalf("expected error for invalid int parsing, got nil")
	}
	if !errors.Is(err, ErrInvalidValue) {
		t.Errorf("expected ErrInvalidValue, got: %v", err)
	}
}

func TestDumpMasked(t *testing.T) {
	cfg := AppConfig{
		AppName: "PaymentService",
		DB: DatabaseConfig{
			Host:     "db.internal",
			Port:     5432,
			Password: "sensitive-plain-text-secret",
		},
	}

	masked := DumpMasked(&cfg)
	if masked["DB_PASSWORD"] != "******" {
		t.Errorf("expected DB_PASSWORD to be '******', got %s", masked["DB_PASSWORD"])
	}
	if masked["NAME"] != "PaymentService" {
		t.Errorf("expected NAME to be 'PaymentService', got %s", masked["NAME"])
	}
}
