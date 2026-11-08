# go-multierror 🧩

> **Thread-Safe Multi-Error Accumulator for Go (Go 1.23+ & Go 1.20+ `Unwrap() []error`).**  
> Form validations, batch worker loops, aur multistep processing me jab ek se zyada errors collect karni hon, tab standard Go `error` interface me unhe wrap karna mushkil hota hai. `go-multierror` Go 1.20+ ke native `errors.Is`, `errors.As`, aur `Unwrap() []error` ko fully support karta hai.

---

## 🚀 Features

- **Standard Go 1.20+ Compatibility**: `errors.Is(merr, TargetErr)` aur `errors.As(merr, &customErr)` pure error collection ko recursively inspect kar sakte hain.
- **Thread-Safe**: Multiple concurrent goroutines bina mutex locks lagaye `merr.Append(err)` call kar sakti hain.
- **Idiomatic `ErrorOrNil()`**: Agar koi error add na hui ho toh `nil` return karta hai, jisse `if err != nil` hamesha accurately behave karta hai.
- **Auto-Flattening**: Ek `*Error` ke andar dusra `*Error` append karne par deep nesting ke bajaye flattened slice banata hai.
- **Filter & Prefix**: Specific types ke errors ko filter karna ya log prefixes add karna (`Worker-1: ...`).

---

## 📦 Quickstart

```go
package main

import (
	"errors"
	"fmt"
	"go-multierror"
)

var (
	ErrInvalidEmail = errors.New("invalid email format")
	ErrShortPassword = errors.New("password must be at least 8 chars")
)

func validateUser(email, pass string) error {
	var merr multierror.Error

	if email == "" {
		merr.Append(ErrInvalidEmail)
	}
	if len(pass) < 8 {
		merr.Append(ErrShortPassword)
	}

	return merr.ErrorOrNil()
}

func main() {
	err := validateUser("", "123")
	if err != nil {
		fmt.Println("Validation Errors:\n", err)

		// errors.Is works transparently!
		if errors.Is(err, ErrInvalidEmail) {
			fmt.Println("👉 Email needs correction")
		}
	}
}
```
