package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/google/uuid"
)

// GenerateOTP membangkitkan 6-digit angka acak aman (misal: "847291")
func GenerateOTP() string {
	maxNum := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, maxNum)
	if err != nil {
		return "123456" // cadangan fallback
	}
	return fmt.Sprintf("%06d", n.Int64())
}

// GenerateResetToken menghasilkan string token acak unik untuk verifikasi reset password
func GenerateResetToken() string {
	return uuid.New().String()
}
