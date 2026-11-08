# go-mockdb 🗄️

> **Lightweight In-Memory SQL Driver Mock for Unit Testing in Go (Go 1.23+).**  
> Clean architecture ke Repository layers ko unit test karne ke liye Docker ya real database chalane ki zarurat nahi hai. `go-mockdb` Go ke native `database/sql/driver` interface ko implement karta hai aur actual `*sql.DB` object provide karta hai jisme queries, transactions aur rows mock kiye ja sakte hain.

---

## 🚀 Features

- **Standard `*sql.DB` Compatibility**: Production repository structs ko bina kisi code change ke actual `*sql.DB` pass kar sakte hain.
- **Regex Query Matching**: `ExpectQuery` aur `ExpectExec` me flexible regex SQL pattern matching.
- **Strict Argument Checking**: `WithArgs(...)` se parameter values ($1, $2) verify hoti hain.
- **Transaction Lifecycle**: `ExpectBegin`, `ExpectCommit`, aur `ExpectRollback` validation.
- **`ExpectationsWereMet`**: Test ke end me confirm karta hai ki koi unexpected query execute nahi hui aur saari expected queries call hui hain.
- **Zero External Dependencies**: Pure Go Standard Library!

---

## 📦 Quickstart

```go
package main

import (
	"fmt"
	"go-mockdb"
)

func main() {
	// 1. Mock DB initialize karein
	db, mock, _ := mockdb.New()
	defer db.Close()

	// 2. Expectations set karein
	mockRows := mockdb.NewRows([]string{"id", "username"}).
		AddRow(1, "arjun").
		AddRow(2, "vikram")

	mock.ExpectQuery(`SELECT id, username FROM users`).
		WillReturnRows(mockRows)

	// 3. Normal database/sql code execute karein
	rows, err := db.Query(`SELECT id, username FROM users`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		_ = rows.Scan(&id, &name)
		fmt.Printf("User %d: %s\n", id, name)
	}

	// 4. Verify all expectations met
	if err := mock.ExpectationsWereMet(); err != nil {
		fmt.Println("Test failed:", err)
	} else {
		fmt.Println("✅ All SQL expectations satisfied!")
	}
}
```
