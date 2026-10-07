package handler

import (
	"net/http"

	"auth-api/internal/service"

	"github.com/gin-gonic/gin"
)

// UserHandler menangani request HTTP seputar pengguna
type UserHandler struct {
	authService *service.AuthService
}

// NewUserHandler membuat instance UserHandler
func NewUserHandler(authService *service.AuthService) *UserHandler {
	return &UserHandler{authService: authService}
}

// GetProfile mengembalikan data profil pengguna yang sedang login
func (h *UserHandler) GetProfile(c *gin.Context) {
	val, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Identitas pengguna tidak ditemukan pada token!"})
		return
	}

	var userID int64
	switch v := val.(type) {
	case float64:
		userID = int64(v)
	case int64:
		userID = v
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Format ID pengguna tidak valid!"})
		return
	}

	user, err := h.authService.GetUserProfile(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "Akun tidak ditemukan!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Data profil berhasil diambil!",
		"data": gin.H{
			"id":              user.ID,
			"email":           user.Email,
			"whatsapp_number": user.WhatsAppNumber,
			"created_at":      user.CreatedAt,
			"updated_at":      user.UpdatedAt,
		},
	})
}
