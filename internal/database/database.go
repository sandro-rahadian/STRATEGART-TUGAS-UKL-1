package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"  // Driver PostgreSQL / Supabase
	_ "modernc.org/sqlite" // Driver SQLite lokal (cadangan offline)
)

// DB membungkus koneksi database dan nama driver
type DB struct {
	*sql.DB
	Driver string
}

// InitDB menghubungkan ke Supabase PostgreSQL atau fallback ke SQLite lokal
func InitDB(databaseURL string) (*DB, error) {
	var conn *sql.DB
	var driver string
	var err error

	if databaseURL != "" {
		log.Println("🔌 Menghubungkan ke Supabase PostgreSQL...")
		conn, err = sql.Open("postgres", databaseURL)
		if err == nil {
			err = conn.Ping()
		}

		if err != nil {
			log.Printf("⚠️ Gagal konek ke PostgreSQL: %v\n", err)
			log.Println("🔄 Beralih otomatis ke database SQLite lokal (local.db)...")
			conn, err = sql.Open("sqlite", "local.db")
			driver = "sqlite"
		} else {
			driver = "postgres"
			log.Println("✅ Berhasil terhubung ke Supabase PostgreSQL!")
		}
	} else {
		log.Println("ℹ️ DATABASE_URL kosong. Menggunakan database SQLite lokal (local.db)...")
		conn, err = sql.Open("sqlite", "local.db")
		driver = "sqlite"
	}

	if err != nil {
		return nil, fmt.Errorf("gagal membuka database: %w", err)
	}

	dbInstance := &DB{
		DB:     conn,
		Driver: driver,
	}

	// Buat tabel jika belum ada sesuai skema ERD
	if err := dbInstance.AutoMigrate(); err != nil {
		return nil, err
	}

	return dbInstance, nil
}

// Rebind mengubah placeholder query (?) menjadi ($1, $2, ...) jika menggunakan PostgreSQL
func (d *DB) Rebind(query string) string {
	if d.Driver != "postgres" {
		return query
	}

	var sb strings.Builder
	idx := 1
	for _, char := range query {
		if char == '?' {
			sb.WriteString(fmt.Sprintf("$%d", idx))
			idx++
		} else {
			sb.WriteRune(char)
		}
	}
	return sb.String()
}

// AutoMigrate membuat tabel users, otp, dan password_reset sesuai desain ERD
func (d *DB) AutoMigrate() error {
	var sqlUsers, sqlOTP, sqlPasswordReset string

	if d.Driver == "postgres" {
		sqlUsers = `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			whatsapp_number VARCHAR(20) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
		);`

		sqlOTP = `
		CREATE TABLE IF NOT EXISTS otp (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			kode_otp VARCHAR(10) NOT NULL,
			tipe VARCHAR(30) NOT NULL,
			expired_at TIMESTAMP WITH TIME ZONE NOT NULL,
			is_used BOOLEAN DEFAULT FALSE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
		);`

		sqlPasswordReset = `
		CREATE TABLE IF NOT EXISTS password_reset (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token VARCHAR(255) NOT NULL,
			expired_at TIMESTAMP WITH TIME ZONE NOT NULL,
			is_used BOOLEAN DEFAULT FALSE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
		);`
	} else {
		// SQLite DDL
		sqlUsers = `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			whatsapp_number TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL
		);`

		sqlOTP = `
		CREATE TABLE IF NOT EXISTS otp (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			kode_otp TEXT NOT NULL,
			tipe TEXT NOT NULL,
			expired_at DATETIME NOT NULL,
			is_used INTEGER DEFAULT 0 NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);`

		sqlPasswordReset = `
		CREATE TABLE IF NOT EXISTS password_reset (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			token TEXT NOT NULL,
			expired_at DATETIME NOT NULL,
			is_used INTEGER DEFAULT 0 NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);`
	}

	if _, err := d.Exec(sqlUsers); err != nil {
		return fmt.Errorf("gagal membuat tabel users: %w", err)
	}
	if _, err := d.Exec(sqlOTP); err != nil {
		return fmt.Errorf("gagal membuat tabel otp: %w", err)
	}
	if _, err := d.Exec(sqlPasswordReset); err != nil {
		return fmt.Errorf("gagal membuat tabel password_reset: %w", err)
	}

	log.Println("✅ Tabel 'users', 'otp', dan 'password_reset' berhasil diverifikasi & siap digunakan!")
	return nil
}
