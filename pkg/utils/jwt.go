package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken membuat JSON Web Token (JWT) yang berlaku selama 24 jam
func GenerateToken(userID int64, email, whatsapp, secret string) (string, error) {
	klaim := jwt.MapClaims{
		"user_id":         userID,
		"email":           email,
		"whatsapp_number": whatsapp,
		"exp":             time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, klaim)
	return token.SignedString([]byte(secret))
}

// ValidateToken memverifikasi apakah string JWT masih sah dan belum kedaluwarsa
func ValidateToken(tokenString, secret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metode signing token tidak valid")
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("token tidak valid atau sudah kedaluwarsa")
	}

	if klaim, ok := token.Claims.(jwt.MapClaims); ok {
		return klaim, nil
	}

	return nil, fmt.Errorf("gagal membaca isi token")
}
