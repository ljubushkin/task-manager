package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/ljubushkin/task-manager/auth"
	"github.com/ljubushkin/task-manager/database"
	"github.com/ljubushkin/task-manager/date"
	"github.com/ljubushkin/task-manager/tasks"
)

func TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		tasks.AddTaskHandler(w, r)
	case http.MethodGet:
		tasks.GetTaskHandler(w, r)
	case http.MethodPut:
		tasks.EditTaskHandler(w, r)
	case http.MethodDelete:
		tasks.DeleteTaskHandler(w, r)
	default:
		http.Error(w, `{"error":"Invalid request method"}`, http.StatusMethodNotAllowed)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Строка подключения к PostgreSQL
	connStr := "user=postgres dbname=mydb host=localhost sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Ошибка при подключении к базе данных:", err)
	}
	defer db.Close()

	// Создание базы данных (создание таблиц и индексов)
	database.CreateDatabase(db)

	// Передаем подключение к базе данных в другие пакеты
	tasks.DB = db
	auth.DB = db

	http.HandleFunc("/api/signin", auth.SigninHandler)
	http.HandleFunc("/api/signup", auth.SignupHandler)
	http.Handle("/api/nextdate", http.HandlerFunc(date.ApiNextDate))
	http.Handle("/api/task", auth.Auth(http.HandlerFunc(TaskHandler)))
	http.Handle("/api/tasks", auth.Auth(http.HandlerFunc(tasks.GetTasksHandler)))
	http.Handle("/api/task/done", auth.Auth(http.HandlerFunc(tasks.DoneTaskHandler)))

	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./web"))))

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	log.Printf("Server is starting on port %s...\n", port)

	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
