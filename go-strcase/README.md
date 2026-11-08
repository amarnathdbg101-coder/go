# go-strcase

High-performance, zero-dependency string case conversion engine with smart acronym recognition for Go.

---

## Features
- **Smart Acronym Recognition**: Converts `HTTPServer` to `http_server` and `userID` to `user_id` without strange broken acronym splits.
- **Multiple Casing Styles**:
  - `ToSnake`: `user_id`
  - `ToCamel`: `userId`
  - `ToPascal`: `UserId`
  - `ToKebab`: `user-id`
  - `ToScreamingSnake`: `USER_ID`
  - `ToCustom`: Any user-defined delimiter and casing
- **Minimal Allocations**: Direct rune scanner with preallocated builders.
- **Zero External Dependencies**: Pure Go standard library.

---

## Quickstart

```go
package main

import (
	"fmt"

	gostrcase "go-strcase"
)

func main() {
	input := "HTTPServerConfiguration"

	fmt.Println("Snake:    ", gostrcase.ToSnake(input))           // http_server_configuration
	fmt.Println("Camel:    ", gostrcase.ToCamel(input))           // httpServerConfiguration
	fmt.Println("Pascal:   ", gostrcase.ToPascal("api_user_auth"))// ApiUserAuth
	fmt.Println("Kebab:    ", gostrcase.ToKebab(input))           // http-server-configuration
	fmt.Println("Screaming:", gostrcase.ToScreamingSnake(input))  // HTTP_SERVER_CONFIGURATION
}
```

---

## Testing
Run unit tests with full race and count coverage:
```bash
go test -v -count=1 ./...
```
