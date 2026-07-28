package auth

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type Claims struct {
	UserID int `json:"user_id"`

	jwt.RegisteredClaims
}

func CreateClaims(userID int) (claims Claims) {
	claims.UserID = userID
	claims.RegisteredClaims = jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	}
	return
}

func CreateToken(claims Claims, key string) (token string, err error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = t.SignedString([]byte(key))
	if err != nil {
		return
	}

	return

}

func ParseToken(tokenS string, key string) (claims *Claims, err error) {
		claims = &Claims{}

	token, err := jwt.ParseWithClaims(tokenS, claims, func(t *jwt.Token) (any, error) {
		return []byte(key), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return

}
