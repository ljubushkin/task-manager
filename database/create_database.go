package database

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // PostgreSQL driver
)

func CreateDatabase(db *sql.DB) {
	// Создание таблицы пользователей с улучшенной структурой
	createUsersTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		username TEXT PRIMARY KEY,
		password TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.Exec(createUsersTableSQL)
	if err != nil {
		log.Fatal("Error creating users table:", err)
	}

	// Создание индекса для быстрого поиска по username
	createUserIndexSQL := `
	CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
	`
	_, err = db.Exec(createUserIndexSQL)
	if err != nil {
		log.Fatal("Error creating users index:", err)
	}

	// Создание таблицы для расписания (оставляем как было)
	createSchedulerTableSQL := `
	CREATE TABLE IF NOT EXISTS scheduler (
		id SERIAL PRIMARY KEY,
		date VARCHAR(8) CHECK (LENGTH(date) = 8),
		title TEXT,
		comment TEXT,
		"repeat" VARCHAR(128)
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

	log.Println("All tables and indexes created successfully")
}
