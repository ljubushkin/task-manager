package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// Credentials хранит данные пользователя для аутентификации
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Claims содержит данные JWT токена
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

var (
	DB     *sql.DB
	jwtKey = []byte(getJWTSecret()) // Секретный ключ из переменных окружения
)

// getJWTSecret возвращает секретный ключ из переменных окружения
func getJWTSecret() string {
	secret := os.Getenv("TODO_JWT_SECRET_KEY")
	if secret == "" {
		log.Fatal("TODO_JWT_SECRET_KEY environment variable not set")
	}
	return secret
}

// Auth middleware для проверки JWT токена
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем и проверяем заголовок Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			sendError(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Разбираем заголовок на части
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			sendError(w, "Invalid authorization format. Expected: Bearer <token>", http.StatusUnauthorized)
			return
		}

		tokenStr := tokenParts[1]
		claims := &Claims{}

		// Парсим токен с проверкой подписи
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			sendError(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Проверяем срок действия токена
		if time.Now().Unix() > claims.ExpiresAt {
			sendError(w, "Token expired", http.StatusUnauthorized)
			return
		}

		// Аутентификация успешна - передаем управление
		next.ServeHTTP(w, r)
	}
}

// SigninHandler обрабатывает запрос на вход
func SigninHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ищем пользователя в базе
	var storedHash string
	err := DB.QueryRow("SELECT password FROM users WHERE username = $1", creds.Username).Scan(&storedHash)
	if err != nil {
		if err == sql.ErrNoRows {
			sendError(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}
		log.Printf("Database error: %v", err)
		sendError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Сравниваем пароли
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(creds.Password)); err != nil {
		sendError(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Генерируем JWT токен
	expirationTime := time.Now().Add(24 * time.Hour) // Токен на 24 часа
	claims := &Claims{
		Username: creds.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Printf("Token generation error: %v", err)
		sendError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token":      tokenString,
		"token_type": "Bearer",
		"expires_in": strconv.FormatInt(expirationTime.Unix(), 10),
	})
}

// SignupHandler обрабатывает регистрацию новых пользователей
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Валидация входных данных
	if creds.Username == "" || creds.Password == "" {
		sendError(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Проверяем существование пользователя
	var exists bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", creds.Username).Scan(&exists)
	if err != nil {
		log.Printf("Database error: %v", err)
		sendError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if exists {
		sendError(w, "Username already exists", http.StatusConflict)
		return
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Password hashing error: %v", err)
		sendError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Сохраняем пользователя
	_, err = DB.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", creds.Username, string(hashedPassword))
	if err != nil {
		log.Printf("User creation error: %v", err)
		sendError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// sendError универсальная функция для отправки ошибок
func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
