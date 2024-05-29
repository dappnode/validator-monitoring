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

type KeyId struct {
	PublicKey string   `json:"publicKey"`
	Tags      []string `json:"tags"`
}

type contextKey string

const TagsKey contextKey = "tags"

// JWTMiddleware dynamically checks tokens against public keys loaded from a JSON file
func JWTMiddleware(next http.Handler, jwtUsersFilePath string) http.Handler {
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

		// Load all key ids from the whitelist JSON data file
		keyIds, err := loadKeyIds(jwtUsersFilePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Internal server error: %v", err), http.StatusInternalServerError)
			return
		}

		// Parse and verify the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, fmt.Errorf("kid not found in token header, generate a new token with a 'kid'")
			}

			// Load the public key for the given kid
			entry, exists := keyIds[kid]
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

		// Extract the kid and find the associated tags. We have to do this again because the token is parsed in a separate function.
		kid, ok := token.Header["kid"].(string)
		if !ok {
			http.Error(w, "kid not found in token header", http.StatusUnauthorized)
			return
		}

		entry, exists := keyIds[kid]
		if !exists {
			http.Error(w, "public key not found for kid", http.StatusUnauthorized)
			return
		}

		// If the key id is found, but no tags are associated with it, it means the key is not authorized to access
		// any signature. This should never happen.
		if len(entry.Tags) == 0 {
			http.Error(w, "no authorized tags found for given kid", http.StatusUnauthorized)
			return
		}

		// Store tags in context. We will use this in the handler to query MongoDB
		ctx := context.WithValue(r.Context(), TagsKey, entry.Tags)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func loadKeyIds(filePath string) (map[string]KeyId, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var keys map[string]KeyId
	if err := json.Unmarshal(data, &keys); err != nil {
		return nil, err
	}

	return keys, nil
}
