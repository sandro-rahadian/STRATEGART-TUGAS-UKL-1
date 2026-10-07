package repository

import (
	"database/sql"
	"fmt"
	"time"

	"auth-api/internal/database"
	"auth-api/internal/model"
)

// PasswordResetRepository mengelola query database untuk tabel PASSWORD_RESET
type PasswordResetRepository struct {
	db *database.DB
}

// NewPasswordResetRepository membuat instance PasswordResetRepository
func NewPasswordResetRepository(db *database.DB) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

// Create mencatat token reset password baru ke tabel password_reset
func (r *PasswordResetRepository) Create(pr *model.PasswordReset) error {
	now := time.Now()
	query := r.db.Rebind(`
		INSERT INTO password_reset (user_id, token, expired_at, is_used, created_at)
		VALUES (?, ?, ?, ?, ?)
	`)

	res, err := r.db.Exec(query, pr.UserID, pr.Token, pr.ExpiredAt, false, now)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err == nil {
		pr.ID = id
	}
	pr.CreatedAt = now
	pr.IsUsed = false
	return nil
}

// FindActiveByToken mencari token reset yang masih aktif (is_used = false)
func (r *PasswordResetRepository) FindActiveByToken(token string) (*model.PasswordReset, error) {
	query := r.db.Rebind(`
		SELECT id, user_id, token, expired_at, is_used, created_at
		FROM password_reset
		WHERE token = ? AND is_used = false
		ORDER BY id DESC LIMIT 1
	`)

	var pr model.PasswordReset
	err := r.db.QueryRow(query, token).Scan(&pr.ID, &pr.UserID, &pr.Token, &pr.ExpiredAt, &pr.IsUsed, &pr.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("token reset password tidak valid atau sudah pernah digunakan")
		}
		return nil, err
	}
	return &pr, nil
}

// MarkUsed mengubah status is_used menjadi true agar token tidak bisa dipakai ulang
func (r *PasswordResetRepository) MarkUsed(resetID int64) error {
	query := r.db.Rebind(`UPDATE password_reset SET is_used = true WHERE id = ?`)
	_, err := r.db.Exec(query, resetID)
	return err
}
