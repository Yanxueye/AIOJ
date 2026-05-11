package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the signed JWT body carried by authenticated requests.
type Claims struct {
	UserID   uint64 `json:"uid"`
	Username string `json:"uname"`
	jwt.RegisteredClaims
}

// JWTManager encapsulates signing/verifying helpers so the secret is not a
// package-level global.
type JWTManager struct {
	secret      []byte
	expireHours int
}

func NewJWTManager(secret string, expireHours int) *JWTManager {
	if expireHours <= 0 {
		expireHours = 72
	}
	return &JWTManager{secret: []byte(secret), expireHours: expireHours}
}

// Sign returns a signed JWT for the given user.
func (m *JWTManager) Sign(userID uint64, username string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(m.expireHours) * time.Hour)),
			Issuer:    "terminaloj",
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(m.secret)
}

// Parse verifies the signature and returns the embedded claims.
func (m *JWTManager) Parse(tokenStr string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
