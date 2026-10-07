package handler

import (
	"net/http"
	"strings"

	"auth-api/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler menangani request HTTP seputar autentikasi
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler membuat instance AuthHandler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// 1. REGISTER
type RegisterRequest struct {
	Email          string `json:"email"`
	WhatsAppNumber string `json:"whatsapp_number"`
	Password       string `json:"password"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Format data JSON tidak valid!"})
		return
	}

	// Validasi input sederhana
	email := strings.TrimSpace(req.Email)
	wa := strings.TrimSpace(req.WhatsAppNumber)
	pass := strings.TrimSpace(req.Password)

	if email == "" || !strings.Contains(email, "@") {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Email harus diisi dengan format yang benar!"})
		return
	}
	if wa == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Nomor WhatsApp (whatsapp_number) wajib diisi!"})
		return
	}
	if len(pass) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Kata sandi minimal 6 karakter!"})
		return
	}

	newUser, err := h.authService.Register(email, wa, pass)
	if err != nil {
		if strings.HasPrefix(err.Error(), "DUPLICATE:") {
			c.JSON(http.StatusConflict, gin.H{"status": "error", "message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Pendaftaran berhasil! Akun Anda telah terdaftar.",
		"data": gin.H{
			"id":              newUser.ID,
			"email":           newUser.Email,
			"whatsapp_number": newUser.WhatsAppNumber,
			"created_at":      newUser.CreatedAt,
		},
	})
}

// 2. LOGIN EMAIL
type LoginEmailRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) LoginEmail(c *gin.Context) {
	var req LoginEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Format data JSON tidak valid!"})
		return
	}

	email := strings.TrimSpace(req.Email)
	pass := strings.TrimSpace(req.Password)

	if email == "" || pass == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Email dan kata sandi wajib diisi!"})
		return
	}

	token, user, err := h.authService.LoginEmail(email, pass)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Login berhasil!",
		"data": gin.H{
			"token":      token,
			"token_type": "Bearer",
			"user": gin.H{
				"id":              user.ID,
				"email":           user.Email,
				"whatsapp_number": user.WhatsAppNumber,
			},
		},
	})
}

// 3A. LOGIN WHATSAPP - MINTA OTP
type RequestWhatsAppOTPRequest struct {
	WhatsAppNumber string `json:"whatsapp_number"`
}

func (h *AuthHandler) RequestWhatsAppOTP(c *gin.Context) {
	var req RequestWhatsAppOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Format data JSON tidak valid!"})
		return
	}

	wa := strings.TrimSpace(req.WhatsAppNumber)
	if wa == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Nomor WhatsApp wajib diisi!"})
		return
	}

	otp, err := h.authService.RequestWhatsAppOTP(wa)
	if err != nil {
		if strings.HasPrefix(err.Error(), "NOT_FOUND:") {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Kode OTP berhasil dikirimkan ke WhatsApp Anda!",
		"data": gin.H{
			"whatsapp_number": wa,
			"berlaku":         "5 Menit",
			"kode_otp":        otp, // Ditampilkan untuk kemudahan pengujian
		},
	})
}

// 3B. LOGIN WHATSAPP - VERIFIKASI OTP
type VerifyWhatsAppOTPRequest struct {
	WhatsAppNumber string `json:"whatsapp_number"`
	KodeOTP        string `json:"kode_otp"`
}

func (h *AuthHandler) VerifyWhatsAppOTP(c *gin.Context) {
	var req VerifyWhatsAppOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Format data JSON tidak valid!"})
		return
	}

	wa := strings.TrimSpace(req.WhatsAppNumber)
	kode := strings.TrimSpace(req.KodeOTP)

	if wa == "" || kode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Nomor WhatsApp dan kode OTP wajib diisi!"})
		return
	}

	token, user, err := h.authService.VerifyWhatsAppOTP(wa, kode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Verifikasi OTP berhasil, Anda telah masuk!",
		"data": gin.H{
			"token":      token,
			"token_type": "Bearer",
			"user": gin.H{
				"id":              user.ID,
				"email":           user.Email,
				"whatsapp_number": user.WhatsAppNumber,
			},
		},
	})
}

// 4A. LUPA PASSWORD - MINTA OTP
type RequestForgotPasswordOTPRequest struct {
	EmailOrWhatsApp string `json:"email_or_whatsapp"`
}

func (h *AuthHandler) RequestForgotPasswordOTP(c *gin.Context) {
	var req RequestForgotPasswordOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Format data JSON tidak valid!"})
		return
	}

	target := strings.TrimSpace(req.EmailOrWhatsApp)
	if target == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Email atau nomor WhatsApp wajib diisi!"})
		return
	}

	otp, err := h.authService.RequestForgotPasswordOTP(target)
	if err != nil {
		if strings.HasPrefix(err.Error(), "NOT_FOUND:") {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Kode OTP reset password berhasil dikirimkan!",
		"data": gin.H{
			"target":   target,
			"berlaku":  "5 Menit",
			"kode_otp": otp,
		},
	})
}

// 4B. LUPA PASSWORD - VERIFIKASI OTP & DAPATKAN TOKEN RESET
type VerifyForgotPasswordOTPRequest struct {
	EmailOrWhatsApp string `json:"email_or_whatsapp"`
	KodeOTP         string `json:"kode_otp"`
}

func (h *AuthHandler) VerifyForgotPasswordOTP(c *gin.Context) {
	var req VerifyForgotPasswordOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Format data JSON tidak valid!"})
		return
	}

	target := strings.TrimSpace(req.EmailOrWhatsApp)
	kode := strings.TrimSpace(req.KodeOTP)

	if target == "" || kode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Email/WhatsApp dan kode OTP wajib diisi!"})
		return
	}

	resetToken, err := h.authService.VerifyForgotPasswordOTP(target, kode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Kode OTP valid! Gunakan reset token ini untuk memasukkan password baru.",
		"data": gin.H{
			"reset_token": resetToken,
			"berlaku":     "15 Menit",
		},
	})
}

// 4C. LUPA PASSWORD - RESET DENGAN TOKEN
type ResetPasswordWithTokenRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func (h *AuthHandler) ResetPasswordWithToken(c *gin.Context) {
	var req ResetPasswordWithTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Format data JSON tidak valid!"})
		return
	}

	token := strings.TrimSpace(req.Token)
	newPass := strings.TrimSpace(req.NewPassword)

	if token == "" || newPass == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Reset token dan kata sandi baru wajib diisi!"})
		return
	}
	if len(newPass) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Kata sandi baru minimal 6 karakter!"})
		return
	}

	err := h.authService.ResetPasswordWithToken(token, newPass)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Kata sandi berhasil diperbarui! Silakan login dengan kata sandi baru Anda.",
	})
}

// 4D. LUPA PASSWORD - RESET LANGSUNG DENGAN OTP (Opsi Cepat)
type ResetPasswordWithOTPRequest struct {
	EmailOrWhatsApp string `json:"email_or_whatsapp"`
	KodeOTP         string `json:"kode_otp"`
	NewPassword     string `json:"new_password"`
}

func (h *AuthHandler) ResetPasswordWithOTP(c *gin.Context) {
	var req ResetPasswordWithOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Format data JSON tidak valid!"})
		return
	}

	target := strings.TrimSpace(req.EmailOrWhatsApp)
	kode := strings.TrimSpace(req.KodeOTP)
	newPass := strings.TrimSpace(req.NewPassword)

	if target == "" || kode == "" || newPass == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Target, kode OTP, dan password baru wajib diisi!"})
		return
	}
	if len(newPass) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Kata sandi baru minimal 6 karakter!"})
		return
	}

	err := h.authService.ResetPasswordWithOTP(target, kode, newPass)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Kata sandi berhasil diperbarui! Silakan login dengan kata sandi baru Anda.",
	})
}
