package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Structured version of Claims Section, as referenced at
// https://tools.ietf.org/html/rfc7519#section-4.1
// See examples for how to use this with your own claim types
type TokenClaim struct {
	User string `json:"username"`
	jwt.StandardClaims
}

// TokenGenerated tokenGenerated
//
// This is used for returning the token
//
// swagger:model TokenGenerated
type TokenGenerated struct {
	// The authentication token
	// A string with the authentication token for the user
	//
	// example: zI1NiIsIsR5cCI6IkpXVCJ9.ezJ1c2VybmFtZSI6ImFkbWluIiwiZXhwIjoxNTI5NTIzNjU0fQ.PPZvRGnR6VA4v7FmgSfQcGQr-VD
	// required: true
	Token string `json:"token"`
	// The expired time for the token
	// A string with the Datetime when the token will be expired
	//
	// example: 2018-06-20 19:40:54.116369887 +0000 UTC m=+43224.838320603
	// required: true
	ExpiredAt string `json:"expired_at"`
}

// GenerateToken generates a signed token with an expiration of <ExpirationSeconds> seconds
func GenerateToken(username string) (TokenGenerated, error) {
	timeExpire := time.Now().Add(time.Second * time.Duration(Get().LoginToken.ExpirationSeconds))
	claim := TokenClaim{
		username,
		jwt.StandardClaims{
			ExpiresAt: timeExpire.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	ss, err := token.SignedString(Get().LoginToken.SigningKey)
	if err != nil {
		return TokenGenerated{}, err
	}

	return TokenGenerated{Token: ss, ExpiredAt: timeExpire.String()}, nil
}

// ValidateToken checks if the input token is still valid
func ValidateToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaim{}, func(token *jwt.Token) (interface{}, error) {
		return Get().LoginToken.SigningKey, nil
	})
	if err != nil {
		return "", err
	}
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return "", fmt.Errorf("Unexpected signing method: %s", token.Header["alg"])
	}
	if token.Valid {
		user := ""
		if sToken, ok := token.Claims.(*TokenClaim); ok {
			user = sToken.User
		}
		return user, nil
	}
	return "", errors.New("Invalid token")
}
