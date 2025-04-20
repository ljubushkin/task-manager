package tasks

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

var DB *sql.DB

func GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Invalid request method"}`, http.StatusMethodNotAllowed)
		return
	}

	const tasksLimit int8 = 50

	search := r.URL.Query().Get("search")
	query := `SELECT id, date, title, comment, repeat FROM scheduler`
	args := []interface{}{}

	if search != "" {

		if parsedDate, err := time.Parse("02.01.2006", search); err == nil {
			query += ` WHERE date = ?`
			args = append(args, parsedDate.Format("20060102"))
		} else {

			query += ` WHERE title LIKE ? OR comment LIKE ?`
			searchTerm := "%" + search + "%"
			args = append(args, searchTerm, searchTerm)
		}
	}

	query += ` ORDER BY date LIMIT ?`
	args = append(args, tasksLimit)

	rows, err := DB.Query(query, args...)
	if err != nil {
		http.Error(w, `{"error":"Failed to retrieve tasks from the database"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tasks := []map[string]string{}
	for rows.Next() {
		var id int
		var date, title, comment, repeat string
		if err := rows.Scan(&id, &date, &title, &comment, &repeat); err != nil {
			http.Error(w, `{"error":"Failed to scan task from the database"}`, http.StatusInternalServerError)
			return
		}

		task := map[string]string{
			"id":      strconv.Itoa(id),
			"date":    date,
			"title":   title,
			"comment": comment,
			"repeat":  repeat,
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, `{"error":"Database error occurred"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"tasks": tasks,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, `{"error":"Failed to encode tasks to JSON"}`, http.StatusInternalServerError)
		return
	}
}
