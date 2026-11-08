# go-diff

Lightweight, reflection-safe structured data diff engine for Go structs, maps, and slices.

---

## Features
- **Deep Hierarchical Diffing**: Recursively traverses nested structs (`user.address.city`), maps, and slices (`tags[2]`).
- **Struct Tag Customization**: Honors `diff:"name"`, falls back to `json:"name"`, and safely ignores sensitive fields with `diff:"-"`.
- **Granular Change Categorization**: Classifies changes cleanly as `ADDED`, `REMOVED`, or `MODIFIED`.
- **Audit-Log Ready**: Generates clean JSON-serializable diff outputs ideal for enterprise compliance trails.
- **Zero External Dependencies**: Pure Go standard library.

---

## Quickstart

```go
package main

import (
	"encoding/json"
	"fmt"

	godiff "go-diff"
)

type Profile struct {
	Name     string `json:"name"`
	Role     string `json:"role"`
	Password string `diff:"-"` // Ignored from diff!
}

func main() {
	oldProfile := Profile{Name: "Amar", Role: "Developer", Password: "old_secret_pwd"}
	newProfile := Profile{Name: "Amar", Role: "Lead Architect", Password: "new_secret_pwd"}

	result, err := godiff.Compare(oldProfile, newProfile)
	if err != nil {
		panic(err)
	}

	if result.HasChanges() {
		fmt.Printf("Detected %d changes:\n", len(result.Changes))
		for _, change := range result.Changes {
			fmt.Printf("- Field '%s': [%s] %v -> %v\n", change.Path, change.Type, change.OldValue, change.NewValue)
		}
	}

	// Output as JSON
	bytes, _ := json.MarshalIndent(result.Changes, "", "  ")
	fmt.Println(string(bytes))
}
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
