package tasks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ljubushkin/task-manager/date"
)

func AddTaskHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var task Task
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&task); err != nil {
		http.Error(w, `{"error":"Invalid request payload"}`, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if task.Title == "" {
		http.Error(w, `{"error":"Task title is required"}`, http.StatusBadRequest)
		return
	}

	if task.Date == "" {
		task.Date = time.Now().Format(date.FormatDate)
	} else {
		_, err := time.Parse(date.FormatDate, task.Date)
		if err != nil {
			http.Error(w, `{"error":"Invalid date format, should be YYYYMMDD"}`, http.StatusBadRequest)
			return
		}
	}

	now := time.Now().Format(date.FormatDate)
	if task.Date < now && task.Repeat == "" {
		task.Date = now
	}

	if task.Repeat != "" {
		parsedDate, err := time.Parse(date.FormatDate, task.Date)
		if err != nil {
			http.Error(w, `{"error":"Invalid date format, should be YYYYMMDD"}`, http.StatusBadRequest)
			return
		}

		nextDate, err := date.NextDate(parsedDate, task.Date, task.Repeat)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
			return
		}

		if nextDate < now {
			nextDate = now
		}

		task.Date = nextDate
	}

	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		http.Error(w, `{"error":"Failed to add task to the database"}`, http.StatusInternalServerError)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, `{"error":"Failed to retrieve task ID"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"id":"%d"}`, id)
}
