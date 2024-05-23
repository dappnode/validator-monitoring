package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	// Define flags for the command-line input
	privateKeyPath := flag.String("private-key", "private.pem", "Path to the RSA private key file")
	subject := flag.String("sub", "user@example.com", "Subject claim for the JWT")
	expiration := flag.String("exp", "24h", "Expiration duration for the JWT (e.g., '24h' for 24 hours)")
	kid := flag.String("kid", "key1", "Key ID (kid) for the JWT")
	flag.Parse()

	// Read the private key file
	privateKeyData, err := os.ReadFile(*privateKeyPath)
	if err != nil {
		log.Fatalf("Failed to read private key file: %v", err)
	}

	// Parse the RSA private key
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	// Parse the expiration duration
	duration, err := time.ParseDuration(*expiration)
	if err != nil {
		log.Fatalf("Failed to parse expiration duration: %v", err)
	}

	// Create a new token object, specifying signing method and claims
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": *subject,
		"exp": time.Now().Add(duration).Unix(),
	})

	// Set the key ID (kid) in the token header
	token.Header["kid"] = *kid

	// Sign the token with the private key
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		log.Fatalf("Failed to sign token: %v", err)
	}

	// Output the token
	fmt.Println(tokenString)
}
