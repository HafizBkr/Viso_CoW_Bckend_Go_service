package utils

import (
	"log"
	"os"

	"github.com/golang-jwt/jwt/v4"
)

func ParseUserFromJWT(tokenString string) (userID, username, role string) {
	secret := os.Getenv("JWT_SECRET")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || token == nil || !token.Valid {
		log.Printf("JWT parse error: %v", err)
		return "", "", ""
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Printf("JWT claims not ok: %+v", token.Claims)
		return "", "", ""
	}
	log.Printf("JWT claims: %+v", claims)
	userID, _ = claims["userId"].(string)
	username, _ = claims["username"].(string)
	role, _ = claims["role"].(string)
	return
}
