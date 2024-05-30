package jwt

import (
	"os"
	"time"

	"github.com/dappnode/validator-monitoring/listener/internal/logger"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(kid, privateKeyPath, subject, expiration string) (string, error) {
	logger.Info("Starting JWT generation")

	privateKeyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		logger.Error("Failed to read private key file: " + err.Error())
		return "", err
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)
	if err != nil {
		logger.Error("Failed to parse private key: " + err.Error())
		return "", err
	}

	claims := jwt.MapClaims{}
	if subject != "" {
		claims["sub"] = subject
		logger.Info("Subject claim set: " + subject)
	}
	if expiration != "" {
		duration, err := time.ParseDuration(expiration)
		if err != nil {
			logger.Error("Failed to parse expiration duration: " + err.Error())
			return "", err
		}
		claims["exp"] = time.Now().Add(duration).Unix()
		logger.Info("Expiration claim set: " + expiration)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	logger.Info("JWT claims prepared")

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		logger.Error("Failed to sign token: " + err.Error())
		return "", err
	}
	logger.Info("JWT generated and signed successfully")

	return tokenString, nil
}
