package jwt

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"log/slog"
	"time"
)

type JWT struct {
	secret []byte
	logger *slog.Logger
}

func NewJTW(secret string, logger *slog.Logger) *JWT {
	return &JWT{[]byte(secret), logger}
}

func (j *JWT) GenerateToken(userID uint, username string) (string, error) {
	claims := jwt.MapClaims{
		"userID":   userID,
		"username": username,
		"exp":      time.Now().Add(time.Minute * 15).Unix(),
	}
	j.logger.Debug("checking claims", "claims:", claims)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		j.logger.Error("error signing token", "err", err)
		return "", err
	}
	return tokenString, nil
}

func (j *JWT) ParseToken(tokenString string) (jwt.MapClaims, error) {
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil || parsedToken == nil {
		j.logger.Error("Error parsing token", slog.String("error", err.Error()))
		return nil, fmt.Errorf("invalid token")
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		j.logger.Debug("from parsed token", "claims", claims)
		return claims, nil
	}
	return nil, err
}
