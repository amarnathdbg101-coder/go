# SPEC: go-data

## 1. Scope & Purpose
Unified data caching, SQL building, pagination, validation, and sanitization suite for Go applications.
Consolidates Caching (Memory + Redis + Singleflight), Cursor/Offset Pagination, SQL Query Builder, Fluent Validator, XSS Sanitizer, URL Slugifier, and Env Loader under one clean, modular package.

## 2. Public API Sub-packages
- `github.com/shopsilo/go-data/cache`: L1 RAM + L2 Redis, Singleflight Dogpile protection, Generics `GetOrSet[T]`
- `github.com/shopsilo/go-data/paginate`: SQL Injection-Proof Cursor & Offset Pagination
- `github.com/shopsilo/go-data/sqlbuilder`: Parameterized SQL Query Builder
- `github.com/shopsilo/go-data/validator`: Reflection-free fluent validator & tag rules
- `github.com/shopsilo/go-data/sanitize`: XSS & SQL LIKE sanitizer

## 3. Concurrency & Memory Guarantees
- Sub-microsecond evaluation speed.
- Singleflight coalesces 1,000 concurrent DB requests into 1 query.
- Thread-safe storage with sync.RWMutex.
