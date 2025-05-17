package tokens

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpredToken  = errors.New("token has expired")
)

type Payload struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (p *Payload) Valid() error {
	if time.Now().After(p.ExpiresAt.Time) {
		return ErrExpredToken
	}
	return nil
}
func NewPayload(username string, duration time.Duration) (*Payload, error) {
	tokenId, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	payload := &Payload{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "Gatekeeper",
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        tokenId.String(),
		},
	}
	return payload, nil
}
