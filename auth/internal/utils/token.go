package utils

import (
	"auth/internal/models"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)


// Create New Token
func GenerateToken(info models.UserInfo, 
					secretKey []byte, duration time.Duration) (string, error) {
	claims := models.UserClaims{
		Username: info.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secretKey)
}


//Verify Token
func VerifyToken(token string, secretKey []byte) (*models.UserClaims, error){
	t, err := jwt.ParseWithClaims(token, 
	&models.UserClaims{},
	func(t *jwt.Token) (any, error) {
		_, ok := t.Method.(*jwt.SigningMethodHMAC)

		if !ok {
			return nil, fmt.Errorf("Invalid token")
		}

		return secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("verify token error")
	}

	claims, ok := t.Claims.(*models.UserClaims)

	if !ok {
		return nil, fmt.Errorf("Invalid token")
	}

	return claims, nil
}
