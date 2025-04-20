package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

func CreateDatabase(db *sql.DB) {

	createUsersTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		username TEXT PRIMARY KEY,
		password TEXT NOT NULL
	);
	`
	_, err := db.Exec(createUsersTableSQL)
	if err != nil {
		log.Fatal("Error creating users table:", err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT CHECK (LENGTH(date) = 8),
		title TEXT,
		comment TEXT,
		repeat TEXT CHECK (LENGTH(repeat) <= 128)
	);
	`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal("Error creating table:", err)
	}

	createIndexSQL := `
	CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);
	`
	_, err = db.Exec(createIndexSQL)
	if err != nil {
		log.Fatal("Error creating index:", err)
	}

	log.Println("Table and index created successfully")
}
