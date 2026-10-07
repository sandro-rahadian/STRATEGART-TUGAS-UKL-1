package model

import "time"

// PasswordReset merepresentasikan entitas tabel PASSWORD_RESET pada database
type PasswordReset struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	ExpiredAt time.Time `json:"expired_at"`
	IsUsed    bool      `json:"is_used"` // bernilai true jika sudah dipakai untuk reset
	CreatedAt time.Time `json:"created_at"`
}
