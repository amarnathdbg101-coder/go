package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ScaffoldProject creates a complete Clean Architecture Go backend project using C:\Library packages.
func ScaffoldProject(targetDir, projectName string) error {
	rootDir := FindLibraryRoot()
	if rootDir == "" {
		return fmt.Errorf("could not locate library workspace; run from the workspace or set LIBKIT_ROOT")
	}

	var projectPath string
	var moduleName string
	if filepath.IsAbs(projectName) {
		projectPath = projectName
		moduleName = filepath.Base(projectName)
	} else {
		projectPath = filepath.Join(targetDir, projectName)
		moduleName = projectName
	}

	if entries, err := os.ReadDir(projectPath); err == nil {
		if len(entries) > 0 {
			return fmt.Errorf("target directory is not empty: %s", projectPath)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect target directory: %w", err)
	}

	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return err
	}

	dirs := []string{
		"cmd/server",
		"config",
		"internal/handler",
		"internal/service",
		"internal/repository",
	}

	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(projectPath, d), 0755); err != nil {
			return err
		}
	}

	// 1. Write go.mod with replace directives to C:\Library
	goModContent := fmt.Sprintf(`module %s

go 1.23

require (
	github.com/amarnathdbg101-coder/go/go-colorlog v0.0.0
	github.com/amarnathdbg101-coder/go/go-env v0.0.0
	github.com/amarnathdbg101-coder/go/go-shutdown v0.0.0
	github.com/amarnathdbg101-coder/go/go-sqlscan v0.0.0
)

replace (
	github.com/amarnathdbg101-coder/go/go-colorlog => %s/go-colorlog
	github.com/amarnathdbg101-coder/go/go-env => %s/go-env
	github.com/amarnathdbg101-coder/go/go-shutdown => %s/go-shutdown
	github.com/amarnathdbg101-coder/go/go-sqlscan => %s/go-sqlscan
)
`, moduleName,
		filepath.ToSlash(rootDir),
		filepath.ToSlash(rootDir),
		filepath.ToSlash(rootDir),
		filepath.ToSlash(rootDir),
	)

	if err := os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte(goModContent), 0644); err != nil {
		return err
	}

	// 2. Write config/config.go
	configContent := fmt.Sprintf(`package config

import (
	"github.com/amarnathdbg101-coder/go/go-env/envvalidator"
	"time"
)

type Config struct {
	Port         int
	DatabaseURL  string
	Environment  string
	WriteTimeout time.Duration
}

func Load() (*Config, error) {
	schema := envvalidator.New()

	schema.Optional("PORT", "8080").AsInt().Between(1024, 65535)
	schema.Optional("DATABASE_URL", "postgres://postgres:password@localhost:5432/%s?sslmode=disable").AsURL()
	schema.Optional("ENV", "development").OneOf("development", "staging", "production")
	schema.Optional("TIMEOUT", "30s").AsDuration()

	env, err := schema.Validate(nil)
	if err != nil {
		return nil, err
	}

	return &Config{
		Port:         env.GetInt("PORT"),
		DatabaseURL:  env.GetString("DATABASE_URL"),
		Environment:  env.GetString("ENV"),
		WriteTimeout: env.GetDuration("TIMEOUT"),
	}, nil
}
`, moduleName)

	if err := os.WriteFile(filepath.Join(projectPath, "config", "config.go"), []byte(configContent), 0644); err != nil {
		return err
	}

	// 3. Write internal/handler/handler.go
	handlerContent := `package handler

import (
	"encoding/json"
	"net/http"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "healthy",
		"code":   http.StatusOK,
	})
}
`
	if err := os.WriteFile(filepath.Join(projectPath, "internal/handler", "handler.go"), []byte(handlerContent), 0644); err != nil {
		return err
	}

	// 4. Write cmd/server/main.go
	mainContent := fmt.Sprintf(`package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"%s/config"
	"%s/internal/handler"

	"github.com/amarnathdbg101-coder/go/go-colorlog"
	goshutdown "github.com/amarnathdbg101-coder/go/go-shutdown"
)

func main() {
	colorlog.Info("Starting service...", "service", "%s")

	// 1. Load validated environment configuration
	cfg, err := config.Load()
	if err != nil {
		colorlog.Error("Configuration error", "err", err)
		return
	}

	// 2. Setup Routing
	mux := http.NewServeMux()
	userHandler := handler.NewUserHandler()
	mux.HandleFunc("/health", userHandler.Health)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: cfg.WriteTimeout,
	}

	// 3. Graceful Shutdown Coordinator
	coordinator := goshutdown.New(10 * time.Second)

	coordinator.RegisterHTTP("http_server", func(ctx context.Context) error {
		colorlog.Info("Stopping HTTP server gracefully...")
		return server.Shutdown(ctx)
	})

	go func() {
		colorlog.Info("Server listening", "port", cfg.Port, "env", cfg.Environment)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			colorlog.Error("Server listen error", "err", err)
		}
	}()

	// Wait for OS termination signal
	if err := coordinator.Wait(); err != nil {
		colorlog.Error("Shutdown error", "err", err)
	}
	colorlog.Info("Service terminated cleanly.")
}
`, moduleName, moduleName, moduleName)

	if err := os.WriteFile(filepath.Join(projectPath, "cmd/server", "main.go"), []byte(mainContent), 0644); err != nil {
		return err
	}

	// 5. Write README.md
	readmeContent := fmt.Sprintf(`# %s 🚀

> **Clean Architecture Microservice powered by C:\Library.**

---

## 🛠️ Quickstart

Run locally:
`+"```powershell"+`
go run cmd/server/main.go
`+"```"+`

Health Check:
`+"```powershell"+`
curl http://localhost:8080/health
`+"```"+`
`, strings.ToUpper(moduleName))

	if err := os.WriteFile(filepath.Join(projectPath, "README.md"), []byte(readmeContent), 0644); err != nil {
		return err
	}

	// 6. Run go mod tidy to ensure immediate runnable out-of-the-box experience
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = projectPath
	tidyCmd.Env = append(os.Environ(), "GOWORK=off")
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy: %w\n%s", err, strings.TrimSpace(string(output)))
	}

	return nil
}
