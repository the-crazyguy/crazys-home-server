package authentication

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TODO: Make it instance-based?
const issuerVal string = "crzyFileServer"

// TODO: Read from file
var signingKey = []byte("demo-signing-key")

type UserClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// TODO: Add more options for expiration/validity?
func NewTokenString(username string) (string, error) {
	if username == "" {
		return "", errors.New("Error cannot be empty")
	}

	c := &UserClaims{
		username,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
			Issuer:    issuerVal,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   username,
		},
	}

	// TODO: USe asymetics signing method later, see https://golang-jwt.github.io/jwt/usage/create/
	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, c)

	return tkn.SignedString(signingKey)
}

func GetClaims(token *jwt.Token) (*UserClaims, error) {
	if token == nil {
		return nil, errors.New("Token cannot be nil")
	}

	claims, ok := token.Claims.(*UserClaims)

	if !ok {
		return nil, errors.New("Unknown claims type")
	}

	if claims.Username == "" {
		return nil, errors.New("Username cannot be empty")
	}

	return claims, nil
}

func ParseToken(tokenStr string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(
		tokenStr,
		&UserClaims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
			}

			return signingKey, nil
		},
	)
}
