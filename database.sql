-- ========================================================
-- DATABASE SCHEMA: REST API AUTENTIKASI (GOLANG & SUPABASE)
-- Sesuai Desain ERD Database: USERS, OTP, PASSWORD_RESET
-- ========================================================

-- 1. TABEL USERS
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    whatsapp_number VARCHAR(20) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Index untuk mempercepat pencarian akun
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_whatsapp ON users(whatsapp_number);

-- 2. TABEL OTP
CREATE TABLE IF NOT EXISTS otp (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kode_otp VARCHAR(10) NOT NULL,
    tipe VARCHAR(30) NOT NULL,                    -- Contoh: 'LOGIN', 'FORGOT_PASSWORD'
    expired_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_used BOOLEAN DEFAULT FALSE NOT NULL,       -- true jika OTP sudah dipakai sekali
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Index pencarian OTP aktif berdasarkan user_id dan tipe
CREATE INDEX IF NOT EXISTS idx_otp_lookup ON otp(user_id, tipe, is_used);

-- 3. TABEL PASSWORD_RESET
CREATE TABLE IF NOT EXISTS password_reset (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL,                  -- Token verifikasi reset password
    expired_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_used BOOLEAN DEFAULT FALSE NOT NULL,       -- true jika token sudah dipakai
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Index pencarian token reset aktif
CREATE INDEX IF NOT EXISTS idx_password_reset_token ON password_reset(token, is_used);
