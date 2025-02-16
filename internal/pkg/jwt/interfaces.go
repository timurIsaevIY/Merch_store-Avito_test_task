package jwt

import "github.com/dgrijalva/jwt-go"

//go:generate mockgen -source=interfaces.go -destination=mocks/mock.go
type JWTInterface interface {
	GenerateToken(userID uint, username string) (string, error)
	ParseToken(tokenString string) (jwt.MapClaims, error)
}
