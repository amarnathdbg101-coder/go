# SPEC: libkit

## 1. Metadata
- **Module**: `libkit`
- **Package**: `main`
- **Location**: `C:\Library\libkit`
- **Binary**: `C:\Library\bin\libkit.exe`
- **Go Version**: `1.23+`
- **Dependencies**: Zero external dependencies (100% Go Standard Library: `os`, `fmt`, `strings`, `path/filepath`, `bufio`, `strconv`, `time`, `sort`, `os/exec`).

---

## 2. Core Capabilities
1. **Catalog Exploration & Discovery**:
   - Parses `C:\Library\CATALOG.md` dynamically with zero external dependencies.
   - Subcommand `list`: Formats all 125 Go libraries in an aligned ANSI terminal table with status indicators.
   - Subcommand `search <query>`: Case-insensitive token and substring search across library names, tags, and architectural responsibilities.
   - Subcommand `info <name|id>`: Pretty prints library details, exported public APIs from `SPEC.md`, and complete quickstart code from `README.md`.
2. **Quality Assurance & Verification**:
   - Subcommand `test <name>`: Automatically runs `go test -v -count=1 ./...` in the targeted library directory.
   - Subcommand `test --all` / `verify`: Iterates through all 125 libraries, runs unit test suites sequentially, collects pass/fail metrics, and reports aggregated execution time and zero-failure guarantees.
3. **Dependency Injection**:
   - Subcommand `add <name> [--to projectDir]`: Safely executes `go mod edit` to inject `require` and local `replace` directives pointing to `C:\Library\<name>` in target Go microservices.
4. **Clean Architecture Project Scaffolder**:
   - Subcommand `new <project-name>`: Generates a complete, idiomatic Clean Architecture Go backend project ready to run with zero setup:
     - `cmd/server/main.go`: HTTP server bootstrap with phased graceful shutdown via `go-shutdown` and structured ANSI logging via `go-colorlog`.
     - `config/config.go`: Startup environment schema validation via `go-envvalidator` with type-safe getters (`PORT`, `DATABASE_URL`, `ENV`, `TIMEOUT`).
     - `internal/handler/`: REST HTTP route handlers with JSON serialization.
     - `internal/service/`: Business domain logic contracts and services.
     - `internal/repository/`: Data persistence layer contracts and database implementations.
     - `go.mod`: Fully initialized with replace directives pointing to `C:\Library` modules and pre-tidied.

---

## 3. Public CLI Interface
```bash
libkit list
libkit search <term>
libkit info <name|id>
libkit test [name|--all]
libkit add <name> [--to projectDir]
libkit new <project-name>
libkit verify
libkit version
```
