# Technical Specification: go-diff

## 1. Domain & Purpose
`go-diff` is a lightweight, reflection-safe structured data difference engine for Go. It calculates detailed, human-readable, and machine-serializable change deltas between two states of structs, maps, slices, and primitive types. It is designed specifically for audit logs, database update tracking, configuration reconciliation, and event-sourcing diffing.

## 2. Architecture & Design Principles
- **Tag-Driven Inspection**: Inspects struct tags `diff:"name"` or falls back to `json:"name"`. Supports `diff:"-"` and `json:"-"` to exclude sensitive fields (e.g. passwords, secrets).
- **Deep Hierarchical Traversal**: Automatically navigates nested structs (`user.address.city`), maps (`config.database.port`), and slices (`items[0]`).
- **Granular Change Types**: Categorizes all differences into `ADDED`, `REMOVED`, or `MODIFIED`.
- **Zero Heavy Dependencies**: Pure Go standard library (`reflect`, `fmt`, `strings`).

## 3. API Surface
- `type ChangeType string` (`Added`, `Removed`, `Modified`)
- `type DiffItem struct { Path string; Type ChangeType; OldValue, NewValue any }`
- `type DiffResult struct { Changes []DiffItem }`
- `func (r *DiffResult) HasChanges() bool`
- `func (r *DiffResult) Added() []DiffItem`
- `func (r *DiffResult) Removed() []DiffItem`
- `func (r *DiffResult) Modified() []DiffItem`
- `func Compare(oldVal, newVal any) (*DiffResult, error)`

## 4. Error Handling
- `ErrTypeMismatch`: Incompatible type comparison without safe conversion path.
- `ErrInvalidInput`: Unparseable input structures.

## 5. Verification & Testing
- Unit tests verify struct field modification, nested struct traversal, sensitive field exclusion (`diff:"-"`), map additions and deletions, and slice index modification/additions.
