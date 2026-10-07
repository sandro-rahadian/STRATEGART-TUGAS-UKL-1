package model

import "time"

// User merepresentasikan entitas tabel USERS pada database
type User struct {
	ID             int64     `json:"id"`
	Email          string    `json:"email"`
	WhatsAppNumber string    `json:"whatsapp_number"`
	PasswordHash   string    `json:"-"` // Tidak ditampilkan saat serialisasi JSON
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
