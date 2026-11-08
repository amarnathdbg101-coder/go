# go-di

Type-safe, generic Dependency Injection container and service locator for Go.

---

## Features
- **Type-Safe Generics**: No type assertions needed (`Resolve[*MyService](c)` returns `*MyService` directly).
- **Flexible Lifetimes**: Supports `Singleton`, `Transient`, and pre-built `Instance` registrations.
- **Named Services**: Differentiate multiple implementations of the same interface (e.g. `primary_db` vs `replica_db`).
- **Circular Dependency Detection**: Prevents infinite recursion and stack overflows.
- **Graceful LIFO Teardown**: `c.Close()` automatically closes connections and singletons implementing `io.Closer` in reverse dependency order.
- **Zero External Dependencies**: Pure Go standard library.

---

## Quickstart

```go
package main

import (
	"fmt"

	godi "go-di"
)

type Database struct {
	DSN string
}

func (d *Database) Close() error {
	fmt.Println("Closing database pool...")
	return nil
}

type UserRepository struct {
	DB *Database
}

type UserService struct {
	Repo *UserRepository
}

func main() {
	c := godi.New()
	defer c.Close()

	// 1. Register Database Singleton
	godi.RegisterSingleton(c, func(cont *godi.Container) (*Database, error) {
		return &Database{DSN: "postgres://localhost:5432/app"}, nil
	})

	// 2. Register Repository (injects Database)
	godi.RegisterSingleton(c, func(cont *godi.Container) (*UserRepository, error) {
		db := godi.MustResolve[*Database](cont)
		return &UserRepository{DB: db}, nil
	})

	// 3. Register Service (injects Repository)
	godi.RegisterSingleton(c, func(cont *godi.Container) (*UserService, error) {
		repo := godi.MustResolve[*UserRepository](cont)
		return &UserService{Repo: repo}, nil
	})

	// 4. Resolve top-level service
	svc := godi.MustResolve[*UserService](c)
	fmt.Printf("Resolved service with DB: %s\n", svc.Repo.DB.DSN)
}
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
