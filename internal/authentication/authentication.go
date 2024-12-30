package authentication

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// TODO: Make it instance-based?
const issuerVal string = "crzyFileServer"

// TODO: Read from file
var signingKey = []byte("demo-signing-key")

type UserClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func NewTokenString(claims *UserClaims) (string, error) {
	if claims.Username == "" {
		return "", errors.New("Username cannot be empty")
	}

	// TODO: USe asymetics signing method later, see https://golang-jwt.github.io/jwt/usage/create/
	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return tkn.SignedString(signingKey)
}

func NewDefaultTokenString(username string) (string, error) {
	c := &UserClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
			Issuer:    issuerVal,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   username,
		},
	}

	return NewTokenString(c)
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

// TODO: Benchmark to get an idea for the optimal cost (10 is the default)
// From a quick search, the general idea is for it to be ~250ms
// Keep in mind it will run on a different machine, so benchmarks should be run
// on the target hardware when deployed
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)

	return string(bytes), err
}

// Returns nil on success, error on failure
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
