package token

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

const minSecuritySize = 32

type JWTMaker struct {
	secretKey string
}

func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecuritySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecuritySize)
	}

	return &JWTMaker{secretKey: secretKey}, nil
}

func (m *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", nil
	}

	// create new jwt token with claims to provide payload with implemented method Valid which checks expire time
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	//finally we generate sighed jwt token by providing secret key
	return jwtToken.SignedString([]byte(m.secretKey))
}
func (m *JWTMaker) VerifyToken(token string) (*Payload, error) {
	//First, we should parse a token:
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, func(token *jwt.Token) (interface{}, error) {
		// token.Method is an interface, and cryptic algorithms use it interface, so in that way below we specify which algorithm we've used.
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}

		return []byte(m.secretKey), nil
	})
	if err != nil {
		validationError, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(validationError.Inner, ErrInvalidToken) {
			return nil, ErrInvalidToken
		}
		return nil, ErrExpiredToken
	}

	//if parse went successfully we can get payload by convert claims into payload object
	// if object satisfies Claims interface we can cast to our payload
	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}

	return payload, nil
}
