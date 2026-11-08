# libkit 🧰

> **Unified Go Library Package Manager & Clean Architecture Scaffolder (Go 1.23+).**  
> `C:\Library` me build kiye gaye sabhi **125 high-performance, zero-dependency Go libraries** ko explore, search, test, inject aur scaffold karne ke liye dedicated CLI tool.

---

## 🚀 Key Features

1. **Catalog Explorer (`list`)**:
   - `C:\Library` ke sabhi 125 libraries ko ANSI colored terminal table me print karta hai.
2. **Instant Search (`search <query>`)**:
   - Kisi bhi keyword (jaise `jwt`, `circuit`, `ratelimit`, `cache`, `sql`, `bloom`, `pool`) se libraries filter karta hai.
3. **Deep Info & Quickstart (`info <name|id>`)**:
   - Target library ki puri summary, `SPEC.md` architecture contract, aur `README.md` quickstart demo seedhe terminal me print karta hai.
4. **Test & Verify (`test [name|--all]`, `verify`)**:
   - Specific library ya fir sabhi 125 libraries ke unit tests run karke verification report deta hai.
5. **Dependency Injection (`add <name>`)**:
   - Kisi bhi existing Go project me seedhe `go mod edit` run karke local `replace` directive aur `require` inject kar deta hai.
6. **Clean Architecture Scaffolder (`new <project-name>`)**:
   - Ek command se complete production-ready Clean Architecture microservice scaffold karta hai (`cmd/server`, `config`, `internal/handler`, `service`, `repository`) jisme Graceful Shutdown (`go-shutdown`), Startup Config Validation (`go-envvalidator`), aur Structured Logging (`go-colorlog`) pre-configured hote hain.
7. **100% Go Standard Library**:
   - Zero external dependencies! Pure Go code.

---

## 📦 Installation & Binary

Binary location:
```powershell
C:\Library\bin\libkit.exe
```

CLI ko system PATH me add karne ke liye (PowerShell):
```powershell
$env:Path += ";C:\Library\bin"
```

---

## 🛠️ Usage Examples

### 1. Catalog List
```powershell
C:\Library\bin\libkit.exe list
```

### 2. Search Libraries
```powershell
# Search for JWT or authentication libraries
C:\Library\bin\libkit.exe search jwt

# Search for Bloom filters
C:\Library\bin\libkit.exe search bloom

# Search for Concurrency or Workers
C:\Library\bin\libkit.exe search worker
```

### 3. Library Information & Code Demo
```powershell
# By name
C:\Library\bin\libkit.exe info go-saga

# By catalog ID
C:\Library\bin\libkit.exe info 116
```

### 4. Run Unit Tests
```powershell
# Test single library
C:\Library\bin\libkit.exe test go-valrules

# Run full batch test suite across all 125 libraries
C:\Library\bin\libkit.exe test --all
```

### 5. Scaffold a New Clean Architecture Backend
```powershell
# Naya project create karo
C:\Library\bin\libkit.exe new my-ecommerce-api

# Project me jao aur run karo
cd my-ecommerce-api
go run cmd/server/main.go
```

Output:
```
[INFO] Starting service... service=my-ecommerce-api
[INFO] Server listening port=8080 env=development
```

Health check:
```powershell
curl http://localhost:8080/health
# {"code":200,"status":"healthy"}
```

### 6. Inject Library into Existing Project
```powershell
cd my-ecommerce-api
C:\Library\bin\libkit.exe add go-valrules
```
Result: `go-valrules` will be added to your `go.mod` with local replace directive to `C:/Library/go-valrules`!

---

## 🏗️ Architecture Design

```
libkit/
├── main.go          # Subcommand CLI parser (list, search, info, test, add, new, verify)
├── catalog.go       # C:\Library\CATALOG.md dynamic parser and metadata cache
├── terminal.go      # ANSI TrueColor renderer, banners, colored status badges, tables
├── scaffold.go      # Clean Architecture code generator and go.mod manager
├── libkit_test.go   # Comprehensive unit test suite
├── SPEC.md          # Architecture and CLI contract
└── README.md        # Hinglish documentation and usage guide
```
