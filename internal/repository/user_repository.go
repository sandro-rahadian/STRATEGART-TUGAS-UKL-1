package repository

import (
	"database/sql"
	"fmt"
	"time"

	"auth-api/internal/database"
	"auth-api/internal/model"
)

// UserRepository mengelola query database untuk tabel USERS
type UserRepository struct {
	db *database.DB
}

// NewUserRepository membuat instance UserRepository
func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create menyimpan pengguna baru ke tabel users
func (r *UserRepository) Create(u *model.User) error {
	now := time.Now()
	query := r.db.Rebind(`
		INSERT INTO users (email, whatsapp_number, password_hash, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?)
	`)

	res, err := r.db.Exec(query, u.Email, u.WhatsAppNumber, u.PasswordHash, now, now)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err == nil {
		u.ID = id
	}
	u.CreatedAt = now
	u.UpdatedAt = now
	return nil
}

// FindByID mencari user berdasarkan ID
func (r *UserRepository) FindByID(id int64) (*model.User, error) {
	query := r.db.Rebind(`
		SELECT id, email, whatsapp_number, password_hash, created_at, updated_at 
		FROM users WHERE id = ? LIMIT 1
	`)

	var u model.User
	err := r.db.QueryRow(query, id).Scan(&u.ID, &u.Email, &u.WhatsAppNumber, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user tidak ditemukan")
		}
		return nil, err
	}
	return &u, nil
}

// FindByEmail mencari user berdasarkan email
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	query := r.db.Rebind(`
		SELECT id, email, whatsapp_number, password_hash, created_at, updated_at 
		FROM users WHERE email = ? LIMIT 1
	`)

	var u model.User
	err := r.db.QueryRow(query, email).Scan(&u.ID, &u.Email, &u.WhatsAppNumber, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user dengan email ini tidak ditemukan")
		}
		return nil, err
	}
	return &u, nil
}

// FindByWhatsApp mencari user berdasarkan nomor WhatsApp
func (r *UserRepository) FindByWhatsApp(whatsappNumber string) (*model.User, error) {
	query := r.db.Rebind(`
		SELECT id, email, whatsapp_number, password_hash, created_at, updated_at 
		FROM users WHERE whatsapp_number = ? LIMIT 1
	`)

	var u model.User
	err := r.db.QueryRow(query, whatsappNumber).Scan(&u.ID, &u.Email, &u.WhatsAppNumber, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("nomor WhatsApp belum terdaftar")
		}
		return nil, err
	}
	return &u, nil
}

// FindByEmailOrWhatsApp mencari user berdasarkan email atau nomor WhatsApp
func (r *UserRepository) FindByEmailOrWhatsApp(identifier string) (*model.User, error) {
	query := r.db.Rebind(`
		SELECT id, email, whatsapp_number, password_hash, created_at, updated_at 
		FROM users WHERE email = ? OR whatsapp_number = ? LIMIT 1
	`)

	var u model.User
	err := r.db.QueryRow(query, identifier, identifier).Scan(&u.ID, &u.Email, &u.WhatsAppNumber, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("akun tidak ditemukan")
		}
		return nil, err
	}
	return &u, nil
}

// UpdatePassword memperbarui kata sandi pengguna
func (r *UserRepository) UpdatePassword(userID int64, newPasswordHash string) error {
	query := r.db.Rebind(`
		UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?
	`)
	_, err := r.db.Exec(query, newPasswordHash, time.Now(), userID)
	return err
}
