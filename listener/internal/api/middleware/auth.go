package middleware

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/dappnode/validator-monitoring/listener/internal/logger"
)

// AuthMiddleware checks if the provided JWT token is valid.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			logger.Debug("No token provided")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// We expect the token to be in the format "Bearer <token>"
		token = strings.TrimPrefix(token, "Bearer ")

		valid, err := isValidToken(token)
		if err != nil || !valid {
			logger.Debug("Invalid token")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isValidToken checks if the provided token exists in the /jwt directory.
func isValidToken(token string) (bool, error) {
	files, err := os.ReadDir("/jwt")
	if err != nil {
		logger.Error("Failed to read /jwt directory: " + err.Error())
		return false, err
	}

	for _, file := range files {
		if !file.IsDir() {
			content, err := os.ReadFile("/jwt/" + file.Name())
			if err != nil {
				logger.Error("Failed to read token file: " + err.Error())
				continue
			}
			if strings.TrimSpace(string(content)) == token {
				return true, nil
			}
		}
	}

	return false, errors.New("token not found")
}
