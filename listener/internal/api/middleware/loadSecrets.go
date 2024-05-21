package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dappnode/validator-monitoring/listener/internal/logger"
	"github.com/golang-jwt/jwt/v5"
)

var allowedPublicKeys []string

// Runs automatically in main.go when the package is imported
func init() {
	// Load public keys from a JSON file
	data, err := os.ReadFile("/app/jwt/public_keys.json")
	if err != nil {
		logger.Fatal("Failed to load public keys: " + err.Error())
	}

	var keys struct {
		AllowedPublicKeys []string `json:"allowedPublicKeys"`
	}

	err = json.Unmarshal(data, &keys)
	if err != nil {
		logger.Fatal("Failed to unmarshal public keys: " + err.Error())
	}

	allowedPublicKeys = keys.AllowedPublicKeys
	logger.Info("Loaded public keys: " + fmt.Sprintln(allowedPublicKeys))
}

// CustomClaims defines the custom claims in the JWT token
type MyCustomClaims struct {
	PubKey string `json:"pubkey"`
	jwt.RegisteredClaims
}

// JWTMiddleware is a middleware that checks the Authorization header for a valid JWT token
// and verifies the signature using the public key of the user contained in the token.
// The public keys are loaded from a JSON file.
// The JWT token must be in the format "Bearer <token string>"
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		// Parse the token
		token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate the algorithm
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, errors.New("unexpected signing method")
			}

			// Extract the claims
			claims, ok := token.Claims.(*MyCustomClaims)
			logger.Info("token:" + fmt.Sprintln(token))
			if !ok {
				return nil, errors.New("invalid token claims")
			}

			// Verify the expiration time
			if claims.ExpiresAt != nil && !claims.ExpiresAt.Time.After(time.Now()) {
				return nil, errors.New("token has expired, it expires at " + claims.ExpiresAt.Time.String() + " now is " + time.Now().String())
			}

			// Verify that the public key is allowed
			isAllowed := false
			for _, allowedKey := range allowedPublicKeys {
				logger.Info("allowedKey:" + allowedKey)
				logger.Info("claims.PubKey:" + claims.PubKey)
				if allowedKey == claims.PubKey {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				return nil, errors.New("public key not allowed")
			}

			// Parse and return the public key
			pubKey, err := jwt.ParseECPublicKeyFromPEM([]byte(claims.PubKey))
			if err != nil {
				return nil, err
			}
			logger.Info("pubKey:" + fmt.Sprintln(pubKey))
			return pubKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
