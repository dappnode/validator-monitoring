package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dappnode/validator-monitoring/listener/internal/jwt"
	"github.com/dappnode/validator-monitoring/listener/internal/logger"
)

func main() {
	privateKeyPath := flag.String("private-key", "", "Path to the RSA private key file (mandatory)")
	subject := flag.String("sub", "", "Subject claim for the JWT (optional)")
	expiration := flag.String("exp", "", "Expiration duration for the JWT in hours (optional)")
	kid := flag.String("kid", "", "Key ID (kid) for the JWT (mandatory)")
	outputFilePath := flag.String("output", "token.jwt", "Output file path for the JWT")

	flag.Parse()

	if *kid == "" || *privateKeyPath == "" {
		logger.Fatal("Key ID (kid) and private key path must be provided")
	}

	tokenString, err := jwt.GenerateJWT(*kid, *privateKeyPath, *subject, *expiration)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Error generating JWT: %v", err))
	}

	fmt.Println("JWT generated successfully:")
	fmt.Println(tokenString)

	err = os.WriteFile(*outputFilePath, []byte(tokenString), 0644)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to write the JWT to file: %v", err))
	}

	fmt.Println("JWT saved to file:", *outputFilePath)
}
