package test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"auth-api/internal/database"
	"auth-api/internal/handler"
	"auth-api/internal/middleware"
	"auth-api/internal/repository"
	"auth-api/internal/service"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

var (
	testRouter *gin.Engine
	testDB     *database.DB
	jwtSecret  = "test-secret-key-12345"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	testDBPath := "test_erd.db"
	_ = os.Remove(testDBPath)

	rawDB, err := sql.Open("sqlite", testDBPath)
	if err != nil {
		panic(err)
	}

	testDB = &database.DB{
		DB:     rawDB,
		Driver: "sqlite",
	}

	if err := testDB.AutoMigrate(); err != nil {
		panic(err)
	}

	// Repositories
	userRepo := repository.NewUserRepository(testDB)
	otpRepo := repository.NewOTPRepository(testDB)
	resetRepo := repository.NewPasswordResetRepository(testDB)

	// Service
	authService := service.NewAuthService(userRepo, otpRepo, resetRepo, jwtSecret)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(authService)

	testRouter = gin.Default()
	api := testRouter.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/login-whatsapp/otp", authHandler.RequestWhatsAppOTP)
			auth.POST("/login-whatsapp/verify", authHandler.VerifyWhatsAppOTP)
			auth.POST("/forgot-password/otp", authHandler.RequestForgotPasswordOTP)
			auth.POST("/forgot-password/verify-otp", authHandler.VerifyForgotPasswordOTP)
			auth.POST("/forgot-password/reset", authHandler.ResetPasswordWithToken)
			auth.POST("/forgot-password/reset-with-otp", authHandler.ResetPasswordWithOTP)
		}

		user := api.Group("/user")
		user.Use(middleware.AuthMiddleware(jwtSecret))
		{
			user.GET("/profile", userHandler.GetProfile)
		}
	}

	exitCode := m.Run()

	_ = testDB.Close()
	_ = os.Remove(testDBPath)

	os.Exit(exitCode)
}

func doRequest(method, url string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}

	req, _ := http.NewRequest(method, url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

func TestAuthFlow(t *testing.T) {
	// 1. REGISTER
	t.Run("Register Sukses", func(t *testing.T) {
		w := doRequest("POST", "/api/auth/register", map[string]string{
			"email":           "budi@example.com",
			"whatsapp_number": "081234567890",
			"password":        "rahasia123",
		}, nil)

		if w.Code != http.StatusCreated {
			t.Fatalf("Ekspektasi 201 Created, dapat: %d", w.Code)
		}
	})

	t.Run("Register Duplikat", func(t *testing.T) {
		w := doRequest("POST", "/api/auth/register", map[string]string{
			"email":           "budi@example.com",
			"whatsapp_number": "081234567890",
			"password":        "rahasia123",
		}, nil)

		if w.Code != http.StatusConflict {
			t.Fatalf("Ekspektasi 409 Conflict, dapat: %d", w.Code)
		}
	})

	// 2. LOGIN (identifier = email ATAU nomor WhatsApp)
	var loginToken string
	t.Run("Login dengan Identifier Email", func(t *testing.T) {
		w := doRequest("POST", "/api/auth/login", map[string]string{
			"identifier": "budi@example.com",
			"password":   "rahasia123",
		}, nil)

		if w.Code != http.StatusOK {
			t.Fatalf("Ekspektasi 200 OK, dapat: %d", w.Code)
		}

		var res struct {
			Data struct {
				Token string `json:"token"`
			} `json:"data"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &res)
		loginToken = res.Data.Token
	})

	t.Run("Login dengan Identifier Nomor WhatsApp", func(t *testing.T) {
		w := doRequest("POST", "/api/auth/login", map[string]string{
			"identifier": "081234567890",
			"password":   "rahasia123",
		}, nil)

		if w.Code != http.StatusOK {
			t.Fatalf("Ekspektasi 200 OK dari login via nomor WhatsApp, dapat: %d", w.Code)
		}
	})

	t.Run("Login dengan Password Salah Ditolak", func(t *testing.T) {
		w := doRequest("POST", "/api/auth/login", map[string]string{
			"identifier": "budi@example.com",
			"password":   "passwordSalahSekali",
		}, nil)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("Ekspektasi 401 Unauthorized, dapat: %d", w.Code)
		}
	})

	// 3. LOGIN WHATSAPP + OTP
	var waOTP string
	t.Run("Minta OTP WhatsApp", func(t *testing.T) {
		w := doRequest("POST", "/api/auth/login-whatsapp/otp", map[string]string{
			"whatsapp_number": "081234567890",
		}, nil)

		if w.Code != http.StatusOK {
			t.Fatalf("Ekspektasi 200 OK, dapat: %d", w.Code)
		}

		var res struct {
			Data struct {
				KodeOTP string `json:"kode_otp"`
			} `json:"data"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &res)
		waOTP = res.Data.KodeOTP
	})

	t.Run("Verifikasi OTP WhatsApp", func(t *testing.T) {
		w := doRequest("POST", "/api/auth/login-whatsapp/verify", map[string]string{
			"whatsapp_number": "081234567890",
			"kode_otp":        waOTP,
		}, nil)

		if w.Code != http.StatusOK {
			t.Fatalf("Ekspektasi 200 OK, dapat: %d", w.Code)
		}
	})

	t.Run("Gagal Pakai Ulang OTP (is_used = true)", func(t *testing.T) {
		w := doRequest("POST", "/api/auth/login-whatsapp/verify", map[string]string{
			"whatsapp_number": "081234567890",
			"kode_otp":        waOTP,
		}, nil)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("Ekspektasi 400 Bad Request saat OTP sudah dipakai, dapat: %d", w.Code)
		}
	})

	// 4. LUPA PASSWORD (OTP -> Reset Token -> Reset Password)
	var resetOTP string
	t.Run("Minta OTP Lupa Password", func(t *testing.T) {
		w := doRequest("POST", "/api/auth/forgot-password/otp", map[string]string{
			"email_or_whatsapp": "budi@example.com",
		}, nil)

		if w.Code != http.StatusOK {
			t.Fatalf("Ekspektasi 200 OK, dapat: %d", w.Code)
		}

		var res struct {
			Data struct {
				KodeOTP string `json:"kode_otp"`
			} `json:"data"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &res)
		resetOTP = res.Data.KodeOTP
	})

	var resetToken string
	t.Run("Verifikasi OTP Lupa Password & Dapatkan Reset Token", func(t *testing.T) {
		w := doRequest("POST", "/api/auth/forgot-password/verify-otp", map[string]string{
			"email_or_whatsapp": "budi@example.com",
			"kode_otp":          resetOTP,
		}, nil)

		if w.Code != http.StatusOK {
			t.Fatalf("Ekspektasi 200 OK, dapat: %d", w.Code)
		}

		var res struct {
			Data struct {
				ResetToken string `json:"reset_token"`
			} `json:"data"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &res)
		resetToken = res.Data.ResetToken
	})

	t.Run("Reset Kata Sandi dengan Token", func(t *testing.T) {
		w := doRequest("POST", "/api/auth/forgot-password/reset", map[string]string{
			"token":        resetToken,
			"new_password": "passwordBaru123",
		}, nil)

		if w.Code != http.StatusOK {
			t.Fatalf("Ekspektasi 200 OK, dapat: %d", w.Code)
		}
	})

	t.Run("Login dengan Password Baru", func(t *testing.T) {
		w := doRequest("POST", "/api/auth/login", map[string]string{
			"identifier": "budi@example.com",
			"password":   "passwordBaru123",
		}, nil)

		if w.Code != http.StatusOK {
			t.Fatalf("Ekspektasi 200 OK, dapat: %d", w.Code)
		}
	})

	// 5. PROTECTED ROUTE PROFILE
	t.Run("Akses Profile dengan Token", func(t *testing.T) {
		w := doRequest("GET", "/api/user/profile", nil, map[string]string{
			"Authorization": "Bearer " + loginToken,
		})

		if w.Code != http.StatusOK {
			t.Fatalf("Ekspektasi 200 OK, dapat: %d", w.Code)
		}
	})

	t.Run("Akses Profile Tanpa Token Ditolak", func(t *testing.T) {
		w := doRequest("GET", "/api/user/profile", nil, nil)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("Ekspektasi 401 Unauthorized, dapat: %d", w.Code)
		}
	})
}
