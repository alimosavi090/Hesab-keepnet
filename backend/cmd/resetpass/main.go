// Command resetpass sets a new password hash for an existing user directly
// in the database. Run on the server where the DB lives:
//
//	go run ./cmd/resetpass /opt/hesab/data/accounting.db cooper 'NewPass123'
package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/ali/hesab-keepnet/backend/internal/passwordhash"
	_ "modernc.org/sqlite"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("usage: resetpass <db-path> <username> <new-password>")
		os.Exit(1)
	}
	dbPath, username, password := os.Args[1], os.Args[2], os.Args[3]
	if len(password) < 8 {
		fmt.Fprintln(os.Stderr, "password must be at least 8 characters")
		os.Exit(1)
	}

	hash, err := passwordhash.Hash(password)
	if err != nil {
		fmt.Fprintln(os.Stderr, "hash:", err)
		os.Exit(1)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open db:", err)
		os.Exit(1)
	}
	defer db.Close()

	res, err := db.Exec(
		"UPDATE users SET password_hash = ?, updated_at = datetime('now') WHERE username = ? AND deleted_at IS NULL",
		hash, username,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "update:", err)
		os.Exit(1)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		// No such user yet — create it (e.g. fresh/restored database).
		res, err := db.Exec(
			"INSERT INTO users (username, password_hash, display_name, role, is_active, created_at, updated_at) VALUES (?, ?, ?, 'ADMIN', 1, datetime('now'), datetime('now'))",
			username, hash, "Administrator",
		)
		if err != nil {
			fmt.Fprintln(os.Stderr, "insert:", err)
			os.Exit(1)
		}
		affected, _ = res.RowsAffected()
	}
	if affected == 0 {
		fmt.Fprintln(os.Stderr, "user not found and insert failed:", username)
		os.Exit(1)
	}
	fmt.Println("password set for", username)
}
