package utils

import (
	"math-top/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey []byte
var tokenExpire time.Duration

type MyClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func init() {
	if config.GlobalConfig == nil {
		config.LoadConfig()
	}
	jwtKey = []byte(config.GlobalConfig.JWT.Secret)
	tokenExpire = time.Duration(config.GlobalConfig.JWT.ExpireHours) * time.Hour
}

func GenerateToken(userID uint, username string, role string) (string, error) {
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

func ParseToken(tokenString string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}
