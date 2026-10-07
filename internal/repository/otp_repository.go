package repository

import (
	"database/sql"
	"fmt"
	"time"

	"auth-api/internal/database"
	"auth-api/internal/model"
)

// OTPRepository mengelola query database untuk tabel OTP
type OTPRepository struct {
	db *database.DB
}

// NewOTPRepository membuat instance OTPRepository
func NewOTPRepository(db *database.DB) *OTPRepository {
	return &OTPRepository{db: db}
}

// Create menyimpan kode OTP baru ke tabel otp
func (r *OTPRepository) Create(o *model.OTP) error {
	now := time.Now()
	query := r.db.Rebind(`
		INSERT INTO otp (user_id, kode_otp, tipe, expired_at, is_used, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`)

	res, err := r.db.Exec(query, o.UserID, o.KodeOTP, o.Tipe, o.ExpiredAt, false, now)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err == nil {
		o.ID = id
	}
	o.CreatedAt = now
	o.IsUsed = false
	return nil
}

// FindLatestActive mencari record OTP terbaru yang belum terpakai (is_used = false)
func (r *OTPRepository) FindLatestActive(userID int64, tipe string) (*model.OTP, error) {
	query := r.db.Rebind(`
		SELECT id, user_id, kode_otp, tipe, expired_at, is_used, created_at
		FROM otp
		WHERE user_id = ? AND tipe = ? AND is_used = false
		ORDER BY id DESC LIMIT 1
	`)

	var o model.OTP
	err := r.db.QueryRow(query, userID, tipe).Scan(&o.ID, &o.UserID, &o.KodeOTP, &o.Tipe, &o.ExpiredAt, &o.IsUsed, &o.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("kode OTP belum pernah diminta atau sudah tidak aktif")
		}
		return nil, err
	}
	return &o, nil
}

// MarkUsed mengubah status is_used menjadi true (karena OTP hanya boleh dipakai 1 kali)
func (r *OTPRepository) MarkUsed(otpID int64) error {
	query := r.db.Rebind(`UPDATE otp SET is_used = true WHERE id = ?`)
	_, err := r.db.Exec(query, otpID)
	return err
}
