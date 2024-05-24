package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/dappnode/validator-monitoring/listener/internal/logger"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	// Define flags for the command-line input
	privateKeyPath := flag.String("private-key", "", "Path to the RSA private key file (mandatory)")
	subject := flag.String("sub", "", "Subject claim for the JWT (optional)")
	expiration := flag.String("exp", "", "Expiration duration for the JWT in hours (optional, e.g., '24h' for 24 hours). If no value is provided, the generated token will not expire.")
	kid := flag.String("kid", "", "Key ID (kid) for the JWT (mandatory)")
	outputFilePath := flag.String("output", "token.jwt", "Output file path for the JWT. Defaults to ./token.jwt")

	flag.Parse()

	// Check for mandatory parameters
	if *kid == "" || *privateKeyPath == "" {
		logger.Fatal("Key ID (kid) and private key path must be provided")
	}

	// Read the private key file
	privateKeyData, err := os.ReadFile(*privateKeyPath)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to read private key file: %v", err))
	}

	// Parse the RSA private key
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to parse private key: %v", err))
	}

	// Prepare the claims for the JWT. These are optional
	claims := jwt.MapClaims{}
	if *subject != "" {
		claims["sub"] = *subject
	}
	if *expiration != "" {
		duration, err := time.ParseDuration(*expiration)
		if err != nil {
			logger.Fatal(fmt.Sprintf("Failed to parse expiration duration: %v", err))
		}
		claims["exp"] = time.Now().Add(duration).Unix()
	}

	// Create a new token object, specifying signing method and claims
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Set the key ID (kid) in the token header
	token.Header["kid"] = *kid

	// Sign the token with the private key
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to sign token: %v", err))
	}

	// Output the token to the console
	fmt.Println("JWT generated successfully:")
	fmt.Println(tokenString)

	// Save the token to a file
	err = os.WriteFile(*outputFilePath, []byte(tokenString), 0644)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to write the JWT to file: %v", err))
	}
	fmt.Println("JWT saved to file:", *outputFilePath)
}
