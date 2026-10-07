package model

import "time"

// OTP merepresentasikan entitas tabel OTP pada database
type OTP struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	KodeOTP   string    `json:"kode_otp"`
	Tipe      string    `json:"tipe"` // misal: 'LOGIN', 'FORGOT_PASSWORD'
	ExpiredAt time.Time `json:"expired_at"`
	IsUsed    bool      `json:"is_used"` // bernilai true jika sudah dipakai sekali
	CreatedAt time.Time `json:"created_at"`
}
