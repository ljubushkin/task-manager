package tasks

import (
	"fmt"
	"net/http"
	"strconv"
)

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `{"error":"Invalid request method"}`, http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, `{"error":"Task ID is required"}`, http.StatusBadRequest)
		return
	}

	_, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, `{"error":"Invalid task ID"}`, http.StatusBadRequest)
		return
	}

	query := `DELETE FROM scheduler WHERE id = $1`
	result, err := DB.Exec(query, id)
	if err != nil {
		http.Error(w, `{"error":"Failed to delete task from the database"}`, http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, `{"error":"Failed to retrieve affected rows"}`, http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, `{"error":"Task not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "{}")
}
