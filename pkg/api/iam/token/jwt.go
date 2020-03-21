package token

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"kubesphere.io/kubesphere/pkg/api/iam"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"time"
)

const DefaultIssuerName = "kubesphere"

var errInvalidToken = errors.New("invalid token")

type claims struct {
	Username string `json:"username"`
	UID      string `json:"uid"`
	// Currently, we are not using any field in jwt.StandardClaims
	jwt.StandardClaims
}

type jwtTokenIssuer struct {
	name    string
	secret  []byte
	keyFunc jwt.Keyfunc
}

func (s *jwtTokenIssuer) Verify(tokenString string) (User, error) {
	if len(tokenString) == 0 {
		return nil, errInvalidToken
	}

	clm := &claims{}

	_, err := jwt.ParseWithClaims(tokenString, clm, s.keyFunc)
	if err != nil {
		return nil, err
	}

	return &iam.User{Username: clm.Username, UID: clm.UID}, nil
}

func (s *jwtTokenIssuer) IssueTo(user User) (string, error) {
	clm := &claims{
		Username: user.GetName(),
		UID:      user.GetUID(),
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			Issuer:    s.name,
			NotBefore: time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, clm)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func NewJwtTokenIssuer(issuerName string, secret []byte) Issuer {
	return &jwtTokenIssuer{
		name:   issuerName,
		secret: secret,
		keyFunc: func(token *jwt.Token) (i interface{}, err error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
				return secret, nil
			} else {
				return nil, fmt.Errorf("expect token signed with HMAC but got %v", token.Header["alg"])
			}
		},
	}
}
