package service

import (
	"fmt"
	"log"
	"time"

	"auth-api/internal/model"
	"auth-api/internal/repository"
	"auth-api/pkg/utils"
)

// AuthService menangani seluruh logika bisnis autentikasi
type AuthService struct {
	userRepo  *repository.UserRepository
	otpRepo   *repository.OTPRepository
	resetRepo *repository.PasswordResetRepository
	jwtSecret string
}

// NewAuthService membuat instance AuthService baru
func NewAuthService(
	userRepo *repository.UserRepository,
	otpRepo *repository.OTPRepository,
	resetRepo *repository.PasswordResetRepository,
	jwtSecret string,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		otpRepo:   otpRepo,
		resetRepo: resetRepo,
		jwtSecret: jwtSecret,
	}
}

// Register mendaftarkan pengguna baru dengan email dan whatsapp_number
func (s *AuthService) Register(email, whatsapp, password string) (*model.User, error) {
	// 1. Cek apakah email sudah terdaftar
	if _, err := s.userRepo.FindByEmail(email); err == nil {
		return nil, fmt.Errorf("DUPLICATE: Email '%s' sudah terdaftar!", email)
	}

	// 2. Cek apakah nomor WhatsApp sudah terdaftar
	if _, err := s.userRepo.FindByWhatsApp(whatsapp); err == nil {
		return nil, fmt.Errorf("DUPLICATE: Nomor WhatsApp '%s' sudah terdaftar!", whatsapp)
	}

	// 3. Enkripsi (Hash) password menggunakan Bcrypt
	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("gagal mengenkripsi kata sandi: %w", err)
	}

	// 4. Simpan ke database
	newUser := &model.User{
		Email:          email,
		WhatsAppNumber: whatsapp,
		PasswordHash:   hash,
	}

	if err := s.userRepo.Create(newUser); err != nil {
		return nil, fmt.Errorf("gagal membuat akun: %w", err)
	}

	log.Printf("🎉 User terdaftar: %s (WA: %s)\n", email, whatsapp)
	return newUser, nil
}

// LoginEmail memvalidasi kredensial email & password dan menghasilkan JWT Token
func (s *AuthService) LoginEmail(email, password string) (string, *model.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return "", nil, fmt.Errorf("email tidak terdaftar atau kata sandi salah")
	}

	// Bandingkan hash password
	if !utils.CheckPassword(password, user.PasswordHash) {
		return "", nil, fmt.Errorf("email tidak terdaftar atau kata sandi salah")
	}

	// Buat token JWT
	token, err := utils.GenerateToken(user.ID, user.Email, user.WhatsAppNumber, s.jwtSecret)
	if err != nil {
		return "", nil, fmt.Errorf("gagal menerbitkan token autentikasi")
	}

	log.Printf("🔑 Login Email berhasil: %s\n", email)
	return token, user, nil
}

// RequestWhatsAppOTP membuat kode OTP untuk login via WhatsApp
func (s *AuthService) RequestWhatsAppOTP(whatsapp string) (string, error) {
	user, err := s.userRepo.FindByWhatsApp(whatsapp)
	if err != nil {
		return "", fmt.Errorf("NOT_FOUND: Nomor WhatsApp belum terdaftar di aplikasi")
	}

	// Bangkitkan 6-digit kode OTP
	kode := utils.GenerateOTP()
	expiredAt := time.Now().Add(5 * time.Minute) // Berlaku 5 menit

	otpRecord := &model.OTP{
		UserID:    user.ID,
		KodeOTP:   kode,
		Tipe:      "LOGIN",
		ExpiredAt: expiredAt,
		IsUsed:    false,
	}

	if err := s.otpRepo.Create(otpRecord); err != nil {
		return "", fmt.Errorf("gagal menyimpan data OTP: %w", err)
	}

	log.Printf("📲 [KIRIM WA KE %s] Kode OTP Login Anda: %s (Berlaku 5 menit)\n", whatsapp, kode)
	return kode, nil
}

// VerifyWhatsAppOTP memverifikasi kode OTP WhatsApp dan mengembalikan JWT Token
func (s *AuthService) VerifyWhatsAppOTP(whatsapp, code string) (string, *model.User, error) {
	user, err := s.userRepo.FindByWhatsApp(whatsapp)
	if err != nil {
		return "", nil, fmt.Errorf("nomor WhatsApp tidak ditemukan")
	}

	// Ambil OTP aktif untuk user ini dengan tipe 'LOGIN'
	otpRecord, err := s.otpRepo.FindLatestActive(user.ID, "LOGIN")
	if err != nil {
		return "", nil, fmt.Errorf("kode OTP belum pernah diminta atau sudah tidak aktif")
	}

	// Periksa apakah waktu OTP sudah kadaluwarsa
	if time.Now().After(otpRecord.ExpiredAt) {
		return "", nil, fmt.Errorf("kode OTP telah kedaluwarsa (lebih dari 5 menit)")
	}

	// Periksa kecocokan kode OTP
	if otpRecord.KodeOTP != code {
		return "", nil, fmt.Errorf("kode OTP yang Anda masukkan tidak valid")
	}

	// Tandai OTP is_used = true (karena OTP hanya boleh dipakai 1 kali!)
	_ = s.otpRepo.MarkUsed(otpRecord.ID)

	// Buat token JWT
	token, err := utils.GenerateToken(user.ID, user.Email, user.WhatsAppNumber, s.jwtSecret)
	if err != nil {
		return "", nil, fmt.Errorf("gagal membuat token autentikasi")
	}

	log.Printf("🔑 Login WhatsApp berhasil: %s\n", whatsapp)
	return token, user, nil
}

// RequestForgotPasswordOTP membuat kode OTP untuk pemulihan password
func (s *AuthService) RequestForgotPasswordOTP(identifier string) (string, error) {
	user, err := s.userRepo.FindByEmailOrWhatsApp(identifier)
	if err != nil {
		return "", fmt.Errorf("NOT_FOUND: Akun dengan email atau nomor WhatsApp tersebut tidak ditemukan")
	}

	kode := utils.GenerateOTP()
	expiredAt := time.Now().Add(5 * time.Minute)

	otpRecord := &model.OTP{
		UserID:    user.ID,
		KodeOTP:   kode,
		Tipe:      "FORGOT_PASSWORD",
		ExpiredAt: expiredAt,
		IsUsed:    false,
	}

	if err := s.otpRepo.Create(otpRecord); err != nil {
		return "", fmt.Errorf("gagal menyimpan OTP lupa password: %w", err)
	}

	log.Printf("🔐 [OTP LUPA PASSWORD] Target %s -> Kode: %s (Berlaku 5 menit)\n", identifier, kode)
	return kode, nil
}

// VerifyForgotPasswordOTP memverifikasi OTP lupa password dan menerbitkan Token Reset di tabel PASSWORD_RESET
func (s *AuthService) VerifyForgotPasswordOTP(identifier, code string) (string, error) {
	user, err := s.userRepo.FindByEmailOrWhatsApp(identifier)
	if err != nil {
		return "", fmt.Errorf("akun tidak ditemukan")
	}

	otpRecord, err := s.otpRepo.FindLatestActive(user.ID, "FORGOT_PASSWORD")
	if err != nil {
		return "", fmt.Errorf("kode OTP belum diminta atau sudah tidak aktif")
	}

	if time.Now().After(otpRecord.ExpiredAt) {
		return "", fmt.Errorf("kode OTP telah kedaluwarsa")
	}

	if otpRecord.KodeOTP != code {
		return "", fmt.Errorf("kode OTP yang dimasukkan tidak valid")
	}

	// Tandai OTP sudah terpakai
	_ = s.otpRepo.MarkUsed(otpRecord.ID)

	// Buat token reset unik dan simpan ke tabel password_reset
	resetToken := utils.GenerateResetToken()
	resetRecord := &model.PasswordReset{
		UserID:    user.ID,
		Token:     resetToken,
		ExpiredAt: time.Now().Add(15 * time.Minute), // Token reset berlaku 15 menit
		IsUsed:    false,
	}

	if err := s.resetRepo.Create(resetRecord); err != nil {
		return "", fmt.Errorf("gagal membuat token reset password")
	}

	log.Printf("🎟️ Token reset password diterbitkan untuk %s\n", identifier)
	return resetToken, nil
}

// ResetPasswordWithToken memperbarui kata sandi menggunakan Token Reset dari tabel PASSWORD_RESET
func (s *AuthService) ResetPasswordWithToken(token, newPassword string) error {
	resetRecord, err := s.resetRepo.FindActiveByToken(token)
	if err != nil {
		return fmt.Errorf("token reset password tidak valid atau sudah kedaluwarsa")
	}

	if time.Now().After(resetRecord.ExpiredAt) {
		return fmt.Errorf("token reset password telah kedaluwarsa (lebih dari 15 menit)")
	}

	// Enkripsi kata sandi baru
	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("gagal mengenkripsi kata sandi baru")
	}

	// Perbarui kata sandi di tabel users
	if err := s.userRepo.UpdatePassword(resetRecord.UserID, hash); err != nil {
		return fmt.Errorf("gagal memperbarui kata sandi: %w", err)
	}

	// Tandai token reset is_used = true (hanya bisa dipakai 1 kali)
	_ = s.resetRepo.MarkUsed(resetRecord.ID)

	log.Printf("🔒 Kata sandi berhasil diperbarui untuk User ID: %d\n", resetRecord.UserID)
	return nil
}

// ResetPasswordWithOTP alur ringkas: langsung reset kata sandi menggunakan OTP
func (s *AuthService) ResetPasswordWithOTP(identifier, code, newPassword string) error {
	resetToken, err := s.VerifyForgotPasswordOTP(identifier, code)
	if err != nil {
		return err
	}
	return s.ResetPasswordWithToken(resetToken, newPassword)
}

// GetUserProfile mengambil profil pengguna berdasarkan User ID
func (s *AuthService) GetUserProfile(userID int64) (*model.User, error) {
	return s.userRepo.FindByID(userID)
}
