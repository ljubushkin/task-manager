package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/ljubushkin/task-manager/auth"
	"github.com/ljubushkin/task-manager/database"
	"github.com/ljubushkin/task-manager/date"
	"github.com/ljubushkin/task-manager/tasks"
	_ "modernc.org/sqlite"
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

	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	tasks.DB = db
	auth.DB = db

	_, err = os.Stat(dbFile)
	if err != nil {
		database.CreateDatabase(tasks.DB)
	} else {
		log.Println("Database already exists")
	}

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
