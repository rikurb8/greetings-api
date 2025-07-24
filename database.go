package main

import (
	"database/sql"
	"fmt"
	"os"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDatabase() error {
	var err error
	dbPath := "./greetings.db"
	
	// Use test database if in test mode
	if os.Getenv("GO_TEST") == "1" || isTestMode() {
		dbPath = "./test_greetings.db"
	}
	
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS measurements (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		temperature REAL,
		humidity REAL,
		moisture REAL
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("error creating measurements table: %w", err)
	}

	return nil
}

func GetDB() *sql.DB {
	return db
}

// Helper to detect if we're in test mode
func isTestMode() bool {
	for _, arg := range os.Args {
		if arg == "-test.v" || arg == "-test.run" {
			return true
		}
	}
	return false
}