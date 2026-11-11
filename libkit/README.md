# libkit 🧰

> **Unified Go Library Catalog & Clean Architecture Scaffolder.**
>
> `C:\Library` workspace ke cataloged Go libraries ko explore, search, test, local development ke liye inject, aur scaffold karne ka CLI tool.

---

## 🚀 Key Features

1. **Catalog Explorer (`list`)**:
   - `CATALOG.md` me listed libraries ko ANSI colored terminal table me print karta hai.
2. **Instant Search (`search <query>`)**:
   - Kisi bhi keyword (jaise `jwt`, `circuit`, `ratelimit`, `cache`, `sql`, `bloom`, `pool`) se libraries filter karta hai.
3. **Deep Info & Quickstart (`info <name|id>`)**:
   - Target library ki puri summary, `SPEC.md` architecture contract, aur `README.md` quickstart demo seedhe terminal me print karta hai.
4. **Test & Verify (`test [name|--all]`, `verify`)**:
   - Ek library ya workspace me mile sabhi Go modules ke tests run karke verification report deta hai.
5. **Dependency Injection (`add <name>`)**:
   - Kisi bhi existing Go project me seedhe `go mod edit` run karke local `replace` directive aur `require` inject kar deta hai.
6. **Clean Architecture Scaffolder (`new <project-name>`)**:
   - Ek command se complete production-ready Clean Architecture microservice scaffold karta hai (`cmd/server`, `config`, `internal/handler`, `service`, `repository`) jisme Graceful Shutdown (`go-shutdown`), Startup Config Validation (`go-env/envvalidator`), aur Structured Logging (`go-colorlog`) pre-configured hote hain.
7. **100% Go Standard Library**:
   - Zero external dependencies! Pure Go code.

---

## 📦 Installation & Binary

Workspace ke `go.work` me declared Go toolchain use karke CLI build karein (currently Go 1.26.5):
```powershell
New-Item -ItemType Directory -Force C:\Library\bin
Set-Location C:\Library\libkit
go build -o C:\Library\bin\libkit.exe .
```

CLI ko current PowerShell session ke `PATH` me add karne ke liye:
```powershell
$env:Path += ";C:\Library\bin"
libkit.exe list
```

Binary `bin` folder me rahe to workspace root automatically mil jata hai. Workspace ko move karne ya binary ko bahar rakhne par `LIBKIT_ROOT` set karein:
```powershell
$env:LIBKIT_ROOT = "D:\path\to\Library"
```

Is setting ko permanent banana ho to Windows **Environment Variables** settings me user-level `LIBKIT_ROOT` aur `PATH` add karein.

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
C:\Library\bin\libkit.exe test go-validator/valrules

# Run tests for every Go module discovered in the workspace
C:\Library\bin\libkit.exe test --all
```

`C:\Library` ek Go workspace hai, single Go module nahi. Isliye repository root par `go test ./...` chalane ke bajay, kisi module folder se test karein ya `libkit test --all` use karein.

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
C:\Library\bin\libkit.exe add go-validator/valrules
```
Ye command local development ke liye `go.mod` me `replace` directive add karta hai. Aisa `go.mod` dusre machine par tabhi chalega jab library same path par ho. Shareable project ke liye local `replace` hata kar published module version use karein:
```powershell
go get github.com/amarnathdbg101-coder/go/go-validator@v1.1.0
```

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
