package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type PublicKeyEntry struct {
	PublicKey string   `json:"publicKey"`
	Roles     []string `json:"roles"`
}

type contextKey string

const RolesKey contextKey = "roles"

// JWTMiddleware dynamically checks tokens against public keys loaded from a JSON file
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Authorization header format must be Bearer {token}", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Load public keys from JSON file
		publicKeys, err := loadPublicKeys("/app/jwt/users.json")
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to load public keys: %v", err), http.StatusInternalServerError)
			return
		}

		// Parse and verify the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, fmt.Errorf("kid not found in token header")
			}

			// Load the public key for the given kid
			entry, exists := publicKeys[kid]
			if !exists {
				return nil, fmt.Errorf("public key not found for kid: %s", kid)
			}

			// Return the public key for signature verification
			return jwt.ParseRSAPublicKeyFromPEM([]byte(entry.PublicKey))
		})

		if err != nil || !token.Valid {
			http.Error(w, fmt.Sprintf("Invalid token or claims: %v", err), http.StatusUnauthorized)
			return
		}

		// Extract the kid and find the associated roles
		kid, ok := token.Header["kid"].(string)
		if !ok {
			http.Error(w, "kid not found in token header", http.StatusUnauthorized)
			return
		}

		entry, exists := publicKeys[kid]
		if !exists {
			http.Error(w, "public key not found for kid", http.StatusUnauthorized)
			return
		}

		// Store roles in context
		ctx := context.WithValue(r.Context(), RolesKey, entry.Roles)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func loadPublicKeys(filePath string) (map[string]PublicKeyEntry, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var keys map[string]PublicKeyEntry
	if err := json.Unmarshal(data, &keys); err != nil {
		return nil, err
	}

	return keys, nil
}
