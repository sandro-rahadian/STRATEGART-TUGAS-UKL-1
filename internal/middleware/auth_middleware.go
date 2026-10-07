package middleware

import (
	"net/http"
	"strings"

	"auth-api/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware memeriksa apakah request memiliki Token JWT yang sah
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Header 'Authorization' tidak ditemukan!",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Format token salah! Gunakan: Bearer <token>",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := utils.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Token tidak valid atau telah kedaluwarsa!",
			})
			c.Abort()
			return
		}

		// Simpan payload user ke dalam context agar bisa diakses oleh handler
		c.Set("user_id", claims["user_id"])
		c.Set("email", claims["email"])
		c.Set("whatsapp_number", claims["whatsapp_number"])

		c.Next()
	}
}
