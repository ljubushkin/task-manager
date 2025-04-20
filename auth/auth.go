package auth

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// DB - глобальная переменная для доступа к базе данных
var DB *sql.DB

// Auth middleware для базовой аутентификации
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			http.Error(w, `{"error":"Требуется авторизация"}`, http.StatusUnauthorized)
			return
		}

		var storedPass string
		err := DB.QueryRow("SELECT password FROM users WHERE username = $1", username).Scan(&storedPass)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, `{"error":"Неверные учетные данные"}`, http.StatusUnauthorized)
				return
			}
			http.Error(w, `{"error":"Ошибка сервера"}`, http.StatusInternalServerError)
			log.Printf("Error checking credentials: %v", err)
			return
		}

		if storedPass != password {
			http.Error(w, `{"error":"Неверные учетные данные"}`, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func SigninHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, `{"error":"Неверный запрос"}`, http.StatusBadRequest)
		return
	}

	var storedPass string
	err := DB.QueryRow("SELECT password FROM users WHERE username = $1", creds.Username).Scan(&storedPass)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Неверный логин или пароль"})
			return
		}
		http.Error(w, `{"error":"Ошибка сервера"}`, http.StatusInternalServerError)
		log.Printf("Error during signin: %v", err)
		return
	}

	if storedPass != creds.Password {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный логин или пароль"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный запрос"})
		log.Printf("Error decoding signup request: %v", err)
		return
	}

	if creds.Username == "" || creds.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Логин и пароль обязательны"})
		return
	}

	// Проверяем существование пользователя
	var exists bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", creds.Username).Scan(&exists)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка сервера"})
		log.Printf("Error checking if user exists: %v", err)
		return
	}

	if exists {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "Пользователь уже существует"})
		return
	}

	// Добавляем нового пользователя
	_, err = DB.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", creds.Username, creds.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка при создании пользователя"})
		log.Printf("Error creating user: %v", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
