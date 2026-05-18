package utils

import (
	"errors"
	"time"

	"drone-management/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Name string      `json:"name"`
	Role domain.Role `json:"role"`
	jwt.RegisteredClaims
}

type JWTSigner struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTSigner(secret string, ttl time.Duration) *JWTSigner {
	return &JWTSigner{secret: []byte(secret), ttl: ttl}
}

func (s *JWTSigner) Issue(principalID uint, name string, role domain.Role, now time.Time) (string, error) {
	claims := Claims{
		Name: name,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   UintToStr(principalID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(s.secret)
}

func (s *JWTSigner) Parse(tokenStr string) (*Claims, error) {
	c := &Claims{}
	tok, err := jwt.ParseWithClaims(tokenStr, c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !tok.Valid {
		return nil, errors.New("invalid token")
	}
	if !c.Role.Valid() {
		return nil, errors.New("invalid role claim")
	}
	return c, nil
}

func UintToStr(n uint) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

func ClaimSubjectUint(c *Claims) (uint, error) {
	var n uint
	for _, ch := range c.Subject {
		if ch < '0' || ch > '9' {
			return 0, errors.New("invalid subject")
		}
		n = n*10 + uint(ch-'0')
	}
	if c.Subject == "" {
		return 0, errors.New("empty subject")
	}
	return n, nil
}
