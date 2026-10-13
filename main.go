package main

import (
	"fmt"
	"log"
	"net/http"

	"auth-api/internal/config"
	"auth-api/internal/database"
	"auth-api/internal/handler"
	"auth-api/internal/middleware"
	"auth-api/internal/repository"
	"auth-api/internal/service"

	"github.com/gin-gonic/gin"
)

// handleSwaggerUI merender halaman visual Swagger UI di browser
func handleSwaggerUI(c *gin.Context) {
	htmlContent := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" />
  <link rel="stylesheet" href="/swagger-dark.css" />
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-themes@1.4.3/themes/dark.css" />
</head>
<body style="background: #1f1f1f; margin: 0;">
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js" crossorigin></script>
<script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-standalone-preset.js" crossorigin></script>
<script>
  window.onload = () => {
    window.ui = SwaggerUIBundle({
      url: '/swagger.json',
      dom_id: '#swagger-ui',
      deepLinking: true,
      presets: [
        SwaggerUIBundle.presets.apis,
        SwaggerUIStandalonePreset
      ],
      layout: "StandaloneLayout"
    });
  };
</script>
</body>
</html>`

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlContent))
}

func main() {
	// 1. Inisialisasi Konfigurasi
	cfg := config.LoadConfig()

	// 2. Inisialisasi Koneksi Database (Supabase PostgreSQL / SQLite fallback)
	db, err := database.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Fatal: Gagal inisialisasi database: %v", err)
	}

	// 3. Inisialisasi Layer Repository (Akses Database)
	userRepo := repository.NewUserRepository(db)
	otpRepo := repository.NewOTPRepository(db)
	resetRepo := repository.NewPasswordResetRepository(db)

	// 4. Inisialisasi Layer Service (Logika Bisnis)
	authService := service.NewAuthService(userRepo, otpRepo, resetRepo, cfg.JWTSecret)

	// 5. Inisialisasi Layer Handler (HTTP Controller)
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(authService)

	// 6. Setup Gin Router
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Dokumentasi Swagger & Postman
	router.GET("/", handleSwaggerUI)
	router.HEAD("/", handleSwaggerUI)
	router.GET("/docs", handleSwaggerUI)
	router.HEAD("/docs", handleSwaggerUI)
	router.GET("/swagger", handleSwaggerUI)
	router.HEAD("/swagger", handleSwaggerUI)
	router.GET("/swagger.json", func(c *gin.Context) { c.File("swagger.json") })
	router.HEAD("/swagger.json", func(c *gin.Context) { c.File("swagger.json") })
	router.GET("/swagger-dark.css", func(c *gin.Context) { c.File("swagger-dark.css") })
	router.HEAD("/swagger-dark.css", func(c *gin.Context) { c.File("swagger-dark.css") })
	router.GET("/postman_collection.json", func(c *gin.Context) { c.File("postman_collection.json") })
	router.HEAD("/postman_collection.json", func(c *gin.Context) { c.File("postman_collection.json") })

	// Group REST API
	api := router.Group("/api")
	{
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/register", authHandler.Register)
			authRoutes.POST("/login", authHandler.Login)
			authRoutes.POST("/login-whatsapp/otp", authHandler.RequestWhatsAppOTP)
			authRoutes.POST("/login-whatsapp/verify", authHandler.VerifyWhatsAppOTP)
			authRoutes.POST("/forgot-password/otp", authHandler.RequestForgotPasswordOTP)
			authRoutes.POST("/forgot-password/verify-otp", authHandler.VerifyForgotPasswordOTP)
			authRoutes.POST("/forgot-password/reset", authHandler.ResetPasswordWithToken)
			authRoutes.POST("/forgot-password/reset-with-otp", authHandler.ResetPasswordWithOTP)
		}

		userRoutes := api.Group("/user")
		userRoutes.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			userRoutes.GET("/profile", userHandler.GetProfile)
		}
	}

	fmt.Println("=======================================================")
	fmt.Printf("🚀 Server REST API Golang aktif di: http://localhost:%s\n", cfg.Port)
	fmt.Printf("📖 Buka Swagger UI di             : http://localhost:%s/swagger\n", cfg.Port)
	fmt.Println("=======================================================")

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("❌ Gagal menjalankan server: %v", err)
	}
}
