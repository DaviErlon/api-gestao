package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/DaviErlon/api-gestao/entities"
	"github.com/DaviErlon/api-gestao/repository"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "user_id"

var secretKey []byte

func init() {
	secretKey = []byte(os.Getenv("key"))
}

func generateToken(user entities.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"name":    user.Name,
		"exp":     time.Now().Add(time.Hour * 2).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func validateToken(tokenStr string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secretKey, nil
	})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	var u entities.User
	err := repository.DB.QueryRow(
		`SELECT id, name, login, password, empresa_id FROM users WHERE login=$1 AND password=$2`,
		creds.Login, creds.Password,
	).Scan(&u.ID, &u.Name, &u.Login, &u.Password, &u.EmpresaID)
	if err != nil {
		http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
		return
	}

	tokenStr, err := generateToken(u)
	if err != nil {
		http.Error(w, "Erro ao gerar token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenStr,
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Token ausente", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Formato inválido", http.StatusUnauthorized)
			return
		}

		token, err := validateToken(parts[1])
		if err != nil || !token.Valid {
			http.Error(w, "Token inválido", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Token inválido", http.StatusUnauthorized)
			return
		}

		// user_id vem como float64 no MapClaims
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "Token inválido", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, int(userIDFloat))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(UserIDKey).(int)
		if !ok || userID != 1 {
			http.Error(w, "Acesso negado", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}