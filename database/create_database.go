package database

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // PostgreSQL driver
)

func CreateDatabase(db *sql.DB) {
	// Создание таблицы пользователей
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

	// Создание таблицы для расписания
	createSchedulerTableSQL := `
	CREATE TABLE IF NOT EXISTS scheduler (
		id SERIAL PRIMARY KEY,
		date VARCHAR(8) CHECK (LENGTH(date) = 8),
		title TEXT,
		comment TEXT,
		"repeat" VARCHAR(128) -- Используем кавычки для зарезервированного слова "repeat"
	);
	`
	_, err = db.Exec(createSchedulerTableSQL)
	if err != nil {
		log.Fatal("Error creating scheduler table:", err)
	}

	// Создание индекса для поля "date"
	createIndexSQL := `
	CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);
	`
	_, err = db.Exec(createIndexSQL)
	if err != nil {
		log.Fatal("Error creating index:", err)
	}

	log.Println("Tables and index created successfully")
}
