# 🚀 REST API Autentikasi (Golang & Supabase)

REST API Autentikasi yang dibangun menggunakan **Golang** dengan **struktur project standar (idiomatic Go project layout)** dan skema database relasional sesuai rancangan ERD: **USERS**, **OTP**, dan **PASSWORD_RESET**.

---

## 📌 Alur Fitur Utama

1. **Register (Pendaftaran Akun)**
   - Input: `email`, `whatsapp_number`, `password`
   - Password dienkripsi aman menggunakan **Bcrypt** sebelum disimpan ke tabel `users`.
   - Mencegah pendaftaran duplikat untuk email dan nomor WhatsApp.
2. **Login Email**
   - Input: `email`, `password`
   - Memvalidasi kecocokan password hash dan mengembalikan **JSON Web Token (JWT)**.
3. **Login WhatsApp + OTP**
   - **Langkah 1 (Minta OTP)**: Masukkan `whatsapp_number` $\rightarrow$ Sistem membuat 6 digit kode OTP acak dan menyimpannya di tabel `otp` (`tipe = 'LOGIN'`, `is_used = false`, berlaku 5 menit).
   - **Langkah 2 (Verifikasi OTP)**: Masukkan `whatsapp_number` & `kode_otp` $\rightarrow$ Sistem memvalidasi kode, mengubah `is_used = true` agar kode tidak bisa dipakai ulang, dan mengembalikan Token JWT.
4. **Lupa Password + OTP & Token Reset**
   - **Langkah 1 (Minta OTP)**: Masukkan `email_or_whatsapp` $\rightarrow$ Sistem mencatat OTP di tabel `otp` (`tipe = 'FORGOT_PASSWORD'`).
   - **Langkah 2 (Verifikasi OTP)**: Masukkan `email_or_whatsapp` & `kode_otp` $\rightarrow$ Jika OTP valid, server menerbitkan token reset unik di tabel `password_reset` (`is_used = false`, berlaku 15 menit).
   - **Langkah 3 (Reset Password)**: Masukkan `token` & `new_password` $\rightarrow$ Sistem memvalidasi token di tabel `password_reset`, mengupdate hash password baru di tabel `users`, dan mengubah `is_used = true`.
   - *(Tersedia juga opsi reset cepat langsung dengan OTP di endpoint `/api/auth/forgot-password/reset-with-otp`)*.
5. **Protected Route (User Profile)**
   - Mengambil data profil akun pengguna login dengan memverifikasi header `Authorization: Bearer <token>`.

---

## 🗂️ Struktur Base Project (Standard Idiomatic Go Layout)

Project ini mengikuti struktur standar resmi bahasa Go:

```text
auth-api/
├── main.go                       # Entry point root (menjalankan server dengan 'go run main.go')
├── cmd/
│   └── api/
│       └── main.go               # Entry point alternatif ('go run ./cmd/api')
├── internal/                     # Kode internal aplikasi (hanya bisa diakses di dalam module ini)
│   ├── config/
│   │   └── config.go             # Membaca environment variables (.env)
│   ├── database/
│   │   └── database.go           # Koneksi DB (Supabase Postgres / SQLite fallback) & migrasi otomatis
│   ├── model/                    # Struct Go representasi tabel database
│   │   ├── user.go               # Struct tabel USERS
│   │   ├── otp.go                # Struct tabel OTP
│   │   └── password_reset.go     # Struct tabel PASSWORD_RESET
│   ├── repository/               # Layer akses data (Query SQL mentah ke database)
│   │   ├── user_repository.go
│   │   ├── otp_repository.go
│   │   └── password_reset_repository.go
│   ├── service/                  # Layer logika bisnis
│   │   └── auth_service.go       # Logika register, login, OTP, token reset, hashing
│   ├── handler/                  # Layer controller HTTP (terima JSON, validasi input, panggil service)
│   │   ├── auth_handler.go
│   │   └── user_handler.go
│   └── middleware/
│       └── auth_middleware.go    # Middleware validasi JWT Bearer token
├── pkg/                          # Paket helper umum yang dapat digunakan kembali
│   └── utils/
│       ├── password.go           # Enkripsi & verifikasi Bcrypt
│       ├── jwt.go                # Pembangkit & validasi token JWT
│       └── otp.go                # Pembangkit 6 digit kode OTP & UUID reset token
├── test/
│   └── api_test.go               # Unit & integration testing otomatis (100% PASS)
├── database.sql                  # Skrip DDL SQL untuk Supabase PostgreSQL
├── postman_collection.json       # File Postman Collection siap import
├── swagger.json                  # Spesifikasi OpenAPI 3.0 / Swagger
├── .env.example                  # Contoh konfigurasi environment
├── go.mod & go.sum               # Dependensi Golang
└── README.md                     # Dokumentasi project
```

---

## 📊 Dokumentasi Database (Sesuai Desain ERD)

### Entity Relationship Diagram (ERD)

```mermaid
erDiagram
    USERS ||--o{ OTP : "memiliki riwayat OTP verifikasi"
    USERS ||--o{ PASSWORD_RESET : "memiliki riwayat token reset password"
    
    USERS {
        int id PK "Primary Key, auto-increment"
        varchar email UQ "Alamat email unik pengguna"
        varchar whatsapp_number UQ "Nomor WhatsApp unik pengguna"
        varchar password_hash "Kata sandi terenkripsi (Bcrypt)"
        datetime created_at "Waktu pembuatan akun"
        datetime updated_at "Waktu pembaruan akun"
    }

    OTP {
        int id PK "Primary Key, auto-increment"
        int user_id FK "Foreign Key merujuk ke USERS.id"
        varchar kode_otp "Kode OTP 6 digit angka"
        varchar tipe "Jenis OTP ('LOGIN' atau 'FORGOT_PASSWORD')"
        datetime expired_at "Batas waktu kedaluwarsa OTP (5 menit)"
        boolean is_used "Status penggunaan: true jika sudah dipakai 1 kali"
        datetime created_at "Waktu pengiriman kode OTP"
    }

    PASSWORD_RESET {
        int id PK "Primary Key, auto-increment"
        int user_id FK "Foreign Key merujuk ke USERS.id"
        varchar token "Token acak unik untuk verifikasi reset"
        datetime expired_at "Batas waktu kedaluwarsa token (15 menit)"
        boolean is_used "Status penggunaan: true jika sudah dipakai reset"
        datetime created_at "Waktu pembuatan token"
    }
```

### Penjelasan Fungsi Kolom `is_used`:
> **`is_used` (Boolean):** Digunakan untuk menjamin prinsip **One-Time Use (OTP)**. Kode OTP atau Token Reset hanya boleh digunakan tepat satu kali. Begitu berhasil diverifikasi, sistem langsung mengubah `is_used = true` sehingga mencegah serangan *replay attack* atau penyalahgunaan kode lama.

---

## 🛠️ Prasyarat & Instalasi

1. Pastikan sudah terpasang **Go (Golang)** versi 1.18+:
   ```bash
   go version
   ```
2. Salin template `.env.example` ke `.env`:
   ```bash
   cp .env.example .env
   ```
   > **💡 Kemudahan untuk Pengujian:** Jika `DATABASE_URL` dikosongkan, server otomatis menggunakan database **SQLite lokal (`local.db`)** sehingga Anda bisa langsung menguji aplikasi secara offline tanpa perlu repot konfigurasi cloud!

---

## ▶️ Cara Menjalankan Project

Jalankan perintah berikut dari folder root:
```bash
go run main.go
# atau:
go run ./cmd/api
```

Terminal akan menampilkan:
```text
=======================================================
🚀 Server REST API Golang aktif di: http://localhost:8080
📖 Buka Swagger UI di             : http://localhost:8080/swagger
=======================================================
```

---

## 🌐 Dokumentasi Swagger UI & Postman Collection

1. **Swagger UI Interaktif:**
   Buka di browser: **[http://localhost:8080/swagger](http://localhost:8080/swagger)** atau **[http://localhost:8080/docs](http://localhost:8080/docs)**  
   Anda dapat langsung mencoba setiap endpoint dengan tombol **"Try it out"** $\rightarrow$ **"Execute"**.
2. **Postman Collection:**
   Import file **`postman_collection.json`** ke aplikasi Postman.  
   - Koleksi sudah dilengkapi *test script* otomatis:
     - Setelah Login, **Token JWT otomatis tersimpan** di variabel environment.
     - Setelah Minta OTP, **Kode OTP otomatis tersimpan** untuk request verifikasi berikutnya.

---

## 🧪 Menjalankan Testing Otomatis

Jalankan perintah:
```bash
go test -v ./...
```
Seluruh skenario pengujian alur utama (*Register*, *Login Email*, *Login WA OTP*, *One-time use validation*, *Lupa Password*, *Reset Token*, dan *Protected Route*) akan berjalan dan menghasilkan status **PASS**.
