package utils

import (
	"math-top/internal/config"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey []byte
var tokenExpire time.Duration
var jwtConfigOnce sync.Once

type MyClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func ensureJWTConfig() {
	jwtConfigOnce.Do(func() {
		if config.GlobalConfig == nil {
			config.LoadConfig()
		}
		jwtKey = []byte(config.GlobalConfig.JWT.Secret)
		tokenExpire = time.Duration(config.GlobalConfig.JWT.ExpireHours) * time.Hour
	})
}

// 加密
func GenerateToken(userID uint, username string, role string) (string, error) {
	ensureJWTConfig()
	claims := MyClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "math-top",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// 解密
func ParseToken(tokenString string) (*MyClaims, error) {
	ensureJWTConfig()
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}
