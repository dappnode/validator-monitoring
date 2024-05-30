package jwt

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testPublicKey string = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArVyO2yhG8kGX14//pdCc
XHEXIGB6qNAuVpiX7go8ZMC7dcSUXizQSxE4ffqegHnNiemyNbHOFAbhFaxszkP7
VOdbYdMe1IoFQ3yGGTfpAZtLWinq/CMwI9CSziSKgif3Rsvbj/xhDfQqfamLqmUJ
Mk8gRdeMp7SiwDHrwVnwUdAgwKioEiTGU2y+C6EHdJeX0I6eFkLvGjbYFt35y0du
wSp0ZYsB+P8glqGQvBOMtvWzoQQ8skuJ8yVBtR+GvU7hPXYBknL4jSLBOzJhHeEW
srsGhn9V5Lo775y7n/ZBJFU0tVn/2zi//HKAVCTfG3J7IHAZqnhEivoM3jaFkzHh
TQIDAQAB
-----END PUBLIC KEY-----`

// Test RSA Private Key (usually generated and safely stored; this is for testing only!)
const testPrivateKeyPath = "../../test/data/private.pem"

// TestGenerateJWT tests the GenerateJWT function. From an example private key in
// ../../test/data/private.pem, it generates a JWT token with the kid "testKid".
func TestGenerateJWT(t *testing.T) {
	kid := "testKid"
	subject := "testSubject"
	expiration := "1h"

	tokenString, err := GenerateJWT(kid, testPrivateKeyPath, subject, expiration)
	if err != nil {
		t.Fatalf("Generating JWT should not produce an error: %v", err)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid not found in token header, generate a new token with a 'kid'")
		}

		if kid != "testKid" {
			return nil, fmt.Errorf("expected kid to be 'testKid', got %v", kid)
		}

		return jwt.ParseRSAPublicKeyFromPEM([]byte(testPublicKey))
	})

	if err != nil {
		t.Fatalf("The token should be valid: %v", err)
	}

	if !token.Valid {
		t.Fatalf("The token should be successfully validated")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("Claims should be of type MapClaims")
	}

	if claims["sub"] != subject {
		t.Fatalf("Expected subject to be %v, got %v", subject, claims["sub"])
	}

	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Add(50*time.Minute).Unix() >= int64(exp) {
			t.Fatalf("Expiration should be correct")
		}
	} else {
		t.Fatalf("Expiration claim missing or not a float64")
	}
}
