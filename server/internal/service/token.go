package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 自签 JWT 的负载。
type Claims struct {
	UserID uint   `json:"uid"`
	OpenID string `json:"openid"`
	jwt.RegisteredClaims
}

// IssueToken 为用户签发 30 天有效的 JWT。
func IssueToken(secret string, userID uint, openID string) (string, error) {
	claims := Claims{
		UserID: userID,
		OpenID: openID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// ParseToken 校验并解析 JWT。
func ParseToken(secret, tokenStr string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}
