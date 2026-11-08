# go-version

SemVer 2.0.0 parser, comparator, and constraint evaluation engine for Go.

---

## Features
- **Strict SemVer 2.0.0 Compliance**: Accurately parses `Major.Minor.Patch`, optional pre-release identifiers, and build metadata.
- **Accurate Precedence**: Evaluates numerical and ASCII pre-release priorities (e.g. `1.0.0 > 1.0.0-rc.1`, `beta > alpha`).
- **Flexible Constraint Matching**:
  - Comparison operators: `>= 1.0.0, < 2.0.0`
  - Caret ranges: `^1.2.3` (compatible with `1.x.x`)
  - Tilde ranges: `~1.2.3` (compatible with `1.2.x`)
- **Ideal for Mobile & SaaS APIs**: Check minimum app client versions for force-updates or API version routing.
- **Zero External Dependencies**: Pure Go standard library.

---

## Quickstart

```go
package main

import (
	"fmt"

	goversion "go-version"
)

func main() {
	// 1. Parse client version
	clientVer := goversion.MustParse("v1.4.2")

	// 2. Minimum version constraint (e.g. forced Flutter update if < 1.5.0)
	minRequired, _ := goversion.ParseConstraint(">= 1.5.0")

	if !minRequired.Check(clientVer) {
		fmt.Println("Client version is too old! Please prompt user to update.")
	} else {
		fmt.Println("Client version is supported.")
	}

	// 3. Caret range matching for plugins
	pluginRule, _ := goversion.ParseConstraint("^2.1.0")
	fmt.Println("Matches 2.4.0? ", pluginRule.CheckString("2.4.0")) // true
	fmt.Println("Matches 3.0.0? ", pluginRule.CheckString("3.0.0")) // false (breaking change)
}
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
