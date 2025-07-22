package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDatabase() error {
	var err error
	db, err = sql.Open("sqlite3", "./greetings.db")
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