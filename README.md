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
        varchar email UK "Alamat email unik pengguna"
        varchar whatsapp_number UK "Nomor WhatsApp unik pengguna"
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

## 🔀 Flowchart Endpoint REST API

Diagram di bawah ini **diambil langsung dari kode sumber** (`main.go`, `internal/handler`, `internal/service`, `internal/repository`, `internal/middleware`), bukan gambar manual. Semua label, status code, dan query database sudah diverifikasi satu per satu.

### Ringkasan Endpoint

| # | Method | Path | Auth | Fungsi | Sukses |
|---|--------|------|------|--------|--------|
| 1 | `POST` | `/api/auth/register` | Publik | Daftar akun baru (email + WhatsApp + password) | `201` |
| 2 | `POST` | `/api/auth/login-email` | Publik | Login pakai email + password | `200` |
| 3 | `POST` | `/api/auth/login-whatsapp/otp` | Publik | Minta kode OTP login WhatsApp | `200` |
| 4 | `POST` | `/api/auth/login-whatsapp/verify` | Publik | Verifikasi OTP login → JWT | `200` |
| 5 | `POST` | `/api/auth/forgot-password/otp` | Publik | Minta OTP lupa password | `200` |
| 6 | `POST` | `/api/auth/forgot-password/verify-otp` | Publik | Verifikasi OTP → terbit reset token | `200` |
| 7 | `POST` | `/api/auth/forgot-password/reset` | Publik | Reset password pakai reset token | `200` |
| 8 | `POST` | `/api/auth/forgot-password/reset-with-otp` | Publik | Reset password langsung pakai OTP (opsi cepat) | `200` |
| 9 | `GET` | `/api/user/profile` | **JWT** | Ambil profil user yang login | `200` |

Endpoint dokumentasi (tidak memakai handler): `GET /`, `GET /docs`, `GET /swagger`, `GET /swagger.json`, `GET /swagger-dark.css`, `GET /postman_collection.json`.

---

**Konvensi diagram di bawah ini:**

| Bentuk | Arti |
|--------|------|
| `([...])` | Awal / akhir request (endpoint) |
| `[...]` | Proses / aksi |
| `{...}` | Percabangan (decision) |
| `[(...)]` | Operasi tulis ke database |
| Warna merah | Response error + HTTP status code |

---

### 1. `POST /api/auth/register` — Pendaftaran Akun

```mermaid
%%{init: {"theme":"default"}}%%
flowchart TD
    A(["POST /api/auth/register"]) --> B["Body JSON<br/>email · whatsapp_number · password"]
    B --> C{"ShouldBindJSON<br/>valid?"}
    C -->|No| E1["400 · Format data JSON tidak valid"]
    C -->|Yes| D["TrimSpace all fields"]
    D --> E{"email not empty<br/>AND contains @ ?"}
    E -->|No| E2["400 · Email harus diisi<br/>dengan format yang benar"]
    E -->|Yes| F{"whatsapp_number<br/>not empty?"}
    F -->|No| E3["400 · Nomor WhatsApp wajib diisi"]
    F -->|Yes| G{"len password<br/>at least 6 ?"}
    G -->|No| E4["400 · Kata sandi minimal 6 karakter"]
    G -->|Yes| H["service.Register"]

    H --> I{"SELECT users<br/>WHERE email ? — found?"}
    I -->|Yes| E5["409 · DUPLICATE<br/>Email sudah terdaftar"]
    I -->|No| J{"SELECT users<br/>WHERE whatsapp_number ? — found?"}
    J -->|Yes| E6["409 · DUPLICATE<br/>Nomor WhatsApp sudah terdaftar"]
    J -->|No| K["utils.HashPassword<br/>Bcrypt"]
    K --> L{"Hash success?"}
    L -->|No| E7["500 · Gagal mengenkripsi<br/>kata sandi"]
    L -->|Yes| M["INSERT INTO users<br/>email · whatsapp_number<br/>password_hash · created_at · updated_at"]
    M --> N{"Insert success?"}
    N -->|No| E8["500 · Gagal membuat akun"]
    N -->|Yes| P[("users row inserted<br/>password_hash = bcrypt")]
    P --> Q(["201 Created<br/>id · email · whatsapp_number · created_at"])
```

---

### 2. `POST /api/auth/login-email` — Login Email + Password

```mermaid
%%{init: {"theme":"default"}}%%
flowchart TD
    A(["POST /api/auth/login-email"]) --> B["Body JSON<br/>email · password"]
    B --> C{"ShouldBindJSON valid?"}
    C -->|No| E1["400 · Format data JSON tidak valid"]
    C -->|Yes| D["TrimSpace both fields"]
    D --> E{"email AND password<br/>both filled?"}
    E -->|No| E2["400 · Email dan kata sandi wajib diisi"]
    E -->|Yes| F["service.LoginEmail"]
    F --> G{"SELECT users<br/>WHERE email ?"}
    G -->|Not found| E3["401 · Email tidak terdaftar<br/>atau kata sandi salah"]
    G -->|Found| H{"Bcrypt CheckPassword<br/>hash match?"}
    H -->|No| E3
    H -->|Yes| I["utils.GenerateToken<br/>HS256 · exp 24 jam<br/>claim: user_id · email · whatsapp_number"]
    I --> J{"Token signed OK?"}
    J -->|No| E4["401 · Gagal menerbitkan<br/>token autentikasi"]
    J -->|Yes| K(["200 OK<br/>token · token_type Bearer · user"])
```

---

### 3. `POST /api/auth/login-whatsapp/otp` — Minta OTP Login WhatsApp

```mermaid
%%{init: {"theme":"default"}}%%
flowchart TD
    A(["POST /api/auth/login-whatsapp/otp"]) --> B["Body JSON<br/>whatsapp_number"]
    B --> C{"ShouldBindJSON valid?"}
    C -->|No| E1["400 · Format data JSON tidak valid"]
    C -->|Yes| D["TrimSpace"]
    D --> E{"whatsapp_number<br/>not empty?"}
    E -->|No| E2["400 · Nomor WhatsApp wajib diisi"]
    E -->|Yes| F["service.RequestWhatsAppOTP"]
    F --> G{"SELECT users<br/>WHERE whatsapp_number ?"}
    G -->|Not found| E3["404 · NOT_FOUND<br/>Nomor WhatsApp belum terdaftar"]
    G -->|Found| H["utils.GenerateOTP<br/>6 digit acak · crypto/rand"]
    H --> I["Set OTP fields<br/>tipe = LOGIN<br/>expired_at = now + 5 menit<br/>is_used = false"]
    I --> J["INSERT INTO otp<br/>user_id · kode_otp · tipe<br/>expired_at · is_used · created_at"]
    J --> K{"Insert success?"}
    K -->|No| E4["500 · Gagal menyimpan data OTP"]
    K -->|Yes| L[("otp row inserted<br/>idx_otp_lookup")]
    L --> M["Log: 📲 KIRIM WA<br/>+ kode OTP"]
    M --> N(["200 OK<br/>whatsapp_number · berlaku 5 Menit · kode_otp"])
```

> ⚠️ **Catatan:** response endpoint ini **mengembalikan `kode_otp` di dalam JSON** (`auth_handler.go:152`) untuk memudahkan pengujian/demo. Pada versi produksi, kode OTP tidak boleh ikut dikembalikan — cukup balas `"OTP berhasil dikirim"`.

---

### 4. `POST /api/auth/login-whatsapp/verify` — Verifikasi OTP Login → JWT

```mermaid
%%{init: {"theme":"default"}}%%
flowchart TD
    A(["POST /api/auth/login-whatsapp/verify"]) --> B["Body JSON<br/>whatsapp_number · kode_otp"]
    B --> C{"ShouldBindJSON valid?"}
    C -->|No| E1["400 · Format data JSON tidak valid"]
    C -->|Yes| D["TrimSpace both fields"]
    D --> E{"Both fields filled?"}
    E -->|No| E2["400 · Nomor WhatsApp dan<br/>kode OTP wajib diisi"]
    E -->|Yes| F["service.VerifyWhatsAppOTP"]
    F --> G{"SELECT users<br/>WHERE whatsapp_number ?"}
    G -->|Not found| E3["400 · Nomor WhatsApp tidak ditemukan"]
    G -->|Found| H["SELECT otp WHERE user_id = ?<br/>AND tipe = LOGIN<br/>AND is_used = false<br/>ORDER BY id DESC LIMIT 1"]
    H -->|No row| E4["400 · Kode OTP belum pernah diminta<br/>atau sudah tidak aktif"]
    H -->|Found| I{"now greater than<br/>expired_at ?"}
    I -->|Yes| E5["400 · Kode OTP telah kedaluwarsa<br/>lebih dari 5 menit"]
    I -->|No| J{"kode_otp matches<br/>stored code?"}
    J -->|No| E6["400 · Kode OTP yang Anda masukkan<br/>tidak valid"]
    J -->|Yes| K["UPDATE otp SET is_used = true<br/>One-time use enforcement"]
    K --> L["utils.GenerateToken<br/>HS256 · exp 24 jam"]
    L --> M(["200 OK<br/>token · token_type Bearer · user"])
```

> 🔐 **Anti-replay:** karena `MarkUsed()` dipanggil **sebelum** token diterbitkan, kode OTP yang sama tidak akan pernah bisa dipakai dua kali — request kedua akan berhenti di langkah `FindLatestActive` dan mendapat `400`.

---

### 5. `POST /api/auth/forgot-password/otp` — Minta OTP Lupa Password

```mermaid
%%{init: {"theme":"default"}}%%
flowchart TD
    A(["POST /api/auth/forgot-password/otp"]) --> B["Body JSON<br/>email_or_whatsapp"]
    B --> C{"ShouldBindJSON valid?"}
    C -->|No| E1["400 · Format data JSON tidak valid"]
    C -->|Yes| D["TrimSpace"]
    D --> E{"identifier not empty?"}
    E -->|No| E2["400 · Email atau nomor<br/>WhatsApp wajib diisi"]
    E -->|Yes| F["service.RequestForgotPasswordOTP"]
    F --> G["SELECT users WHERE email = id<br/>OR whatsapp_number = id"]
    G -->|Not found| E3["404 · NOT_FOUND<br/>Akun dengan email atau nomor<br/>WhatsApp tersebut tidak ditemukan"]
    G -->|Found| H["utils.GenerateOTP<br/>6 digit acak"]
    H --> I["Set OTP fields<br/>tipe = FORGOT_PASSWORD<br/>expired_at = now + 5 menit<br/>is_used = false"]
    I --> J["INSERT INTO otp"]
    J --> K{"Insert success?"}
    K -->|No| E4["500 · Gagal menyimpan<br/>OTP lupa password"]
    K -->|Yes| L[("otp row inserted")]
    L --> M(["200 OK<br/>target · berlaku 5 Menit · kode_otp"])
```

> 📌 Bedanya dengan `POST /api/auth/login-whatsapp/otp`: pencarian user memakai `email OR whatsapp_number`, dan `tipe` OTP diisi `FORGOT_PASSWORD` — sehingga OTP login tidak akan bisa dipakai untuk reset password (dan sebaliknya).

---

### 6. `POST /api/auth/forgot-password/verify-otp` — Verifikasi OTP → Terbitkan Reset Token

```mermaid
%%{init: {"theme":"default"}}%%
flowchart TD
    A(["POST /api/auth/forgot-password/verify-otp"]) --> B["Body JSON<br/>email_or_whatsapp · kode_otp"]
    B --> C{"ShouldBindJSON valid?"}
    C -->|No| E1["400 · Format data JSON tidak valid"]
    C -->|Yes| D["TrimSpace both fields"]
    D --> E{"Both fields filled?"}
    E -->|No| E2["400 · Email/WhatsApp dan<br/>kode OTP wajib diisi"]
    E -->|Yes| F["service.VerifyForgotPasswordOTP"]
    F --> G["SELECT users WHERE email = id<br/>OR whatsapp_number = id"]
    G -->|Not found| E3["400 · Akun tidak ditemukan"]
    G -->|Found| H["SELECT otp WHERE user_id = ?<br/>AND tipe = FORGOT_PASSWORD<br/>AND is_used = false<br/>ORDER BY id DESC LIMIT 1"]
    H -->|No row| E4["400 · Kode OTP belum diminta<br/>atau sudah tidak aktif"]
    H -->|Found| I{"now greater than<br/>expired_at ?"}
    I -->|Yes| E5["400 · Kode OTP telah kedaluwarsa"]
    I -->|No| J{"kode_otp matches?"}
    J -->|No| E6["400 · Kode OTP yang dimasukkan<br/>tidak valid"]
    J -->|Yes| K["UPDATE otp SET is_used = true"]
    K --> L["utils.GenerateResetToken<br/>UUID v4 string"]
    L --> M["Set reset fields<br/>expired_at = now + 15 menit<br/>is_used = false"]
    M --> N["INSERT INTO password_reset<br/>user_id · token<br/>expired_at · is_used · created_at"]
    N --> O{"Insert success?"}
    O -->|No| E7["400 · Gagal membuat<br/>token reset password"]
    O -->|Yes| P[("password_reset row inserted<br/>idx_password_reset_token")]
    P --> Q(["200 OK<br/>reset_token · berlaku 15 Menit"])
```

---

### 7. `POST /api/auth/forgot-password/reset` — Reset Password Pakai Reset Token

```mermaid
%%{init: {"theme":"default"}}%%
flowchart TD
    A(["POST /api/auth/forgot-password/reset"]) --> B["Body JSON<br/>token · new_password"]
    B --> C{"ShouldBindJSON valid?"}
    C -->|No| E1["400 · Format data JSON tidak valid"]
    C -->|Yes| D["TrimSpace both fields"]
    D --> E{"Both fields filled?"}
    E -->|No| E2["400 · Reset token dan<br/>kata sandi baru wajib diisi"]
    E -->|Yes| F{"len new_password<br/>at least 6 ?"}
    F -->|No| E3["400 · Kata sandi baru<br/>minimal 6 karakter"]
    F -->|Yes| G["service.ResetPasswordWithToken"]
    G --> H["SELECT password_reset<br/>WHERE token = ?<br/>AND is_used = false<br/>ORDER BY id DESC LIMIT 1"]
    H -->|No row| E4["400 · Token reset password tidak valid<br/>atau sudah pernah digunakan"]
    H -->|Found| I{"now greater than<br/>expired_at ?"}
    I -->|Yes| E5["400 · Token reset telah kedaluwarsa<br/>lebih dari 15 menit"]
    I -->|No| J["utils.HashPassword<br/>Bcrypt new_password"]
    J --> K["UPDATE users<br/>SET password_hash, updated_at<br/>WHERE id = user_id"]
    K --> L{"Update success?"}
    L -->|No| E6["400 · Gagal memperbarui kata sandi"]
    L -->|Yes| M["UPDATE password_reset<br/>SET is_used = true"]
    M --> N(["200 OK<br/>Kata sandi berhasil diperbarui<br/>silakan login ulang"])
```

> 🔁 **Konsekuensi:** karena `password_hash` berubah, **semua JWT lama milik user tersebut masih valid sampai kedaluwarsa (24 jam)** karena token tidak menyimpan `password_hash`. Untuk keamanan produksi, token sebaiknya berisi `token_version` atau `updated_at` yang dicek di middleware.

---

### 8. `POST /api/auth/forgot-password/reset-with-otp` — Reset Password Langsung (Opsi Cepat)

```mermaid
%%{init: {"theme":"default"}}%%
flowchart TD
    A(["POST /api/auth/forgot-password/reset-with-otp"]) --> B["Body JSON<br/>email_or_whatsapp · kode_otp · new_password"]
    B --> C{"ShouldBindJSON valid?"}
    C -->|No| E1["400 · Format data JSON tidak valid"]
    C -->|Yes| D["TrimSpace all fields"]
    D --> E{"All three fields filled?"}
    E -->|No| E2["400 · Target, kode OTP,<br/>dan password baru wajib diisi"]
    E -->|Yes| F{"len new_password<br/>at least 6 ?"}
    F -->|No| E3["400 · Kata sandi baru<br/>minimal 6 karakter"]
    F -->|Yes| G["service.ResetPasswordWithOTP"]

    G --> H["STEP 1<br/>Reuse VerifyForgotPasswordOTP<br/>email_or_whatsapp, kode_otp"]
    H --> I{"OTP valid?"}
    I -->|No| E4["400 · Error diteruskan dari<br/>langkah verifikasi OTP"]
    I -->|Yes| J["UPDATE otp SET is_used = true<br/>Generate UUID reset_token<br/>INSERT password_reset<br/>expired_at = now + 15 menit"]

    J --> K["STEP 2<br/>Reuse ResetPasswordWithToken<br/>reset_token, new_password"]
    K --> L{"Token still active<br/>and not expired?"}
    L -->|No| E5["400 · Token reset tidak valid<br/>atau telah kedaluwarsa"]
    L -->|Yes| M["Bcrypt hash new_password<br/>UPDATE users password_hash<br/>UPDATE password_reset is_used = true"]
    M --> N(["200 OK<br/>Kata sandi berhasil diperbarui"])
```

> ♻️ Endpoint ini **tidak punya logika sendiri** — `ResetPasswordWithOTP()` (`auth_service.go:250`) hanya mengkomposisi `VerifyForgotPasswordOTP()` lalu `ResetPasswordWithToken()` dalam satu request. Diagram di atas menunjukkan kedua subroutine yang dipanggil, jadi alur database-nya identik dengan `forgot-password/verify-otp` (bagian 6) + `forgot-password/reset` (bagian 7) digabung.

---

### 9. `GET /api/user/profile` — Protected Route (JWT Bearer)

```mermaid
%%{init: {"theme":"default"}}%%
flowchart TD
    A(["GET /api/user/profile<br/>Header: Authorization"]) --> B["middleware.AuthMiddleware"]
    B --> C{"Header Authorization<br/>present?"}
    C -->|Missing| E1["401 · Header Authorization<br/>tidak ditemukan · Abort"]
    C -->|Present| D{"Split into 2 parts<br/>AND first part = Bearer ?"}
    D -->|Invalid| E2["401 · Format token salah<br/>Gunakan Bearer + token · Abort"]
    D -->|Valid| F["utils.ValidateToken<br/>HS256 signature check<br/>exp check"]
    F -->|Invalid / expired| E3["401 · Token tidak valid<br/>atau telah kedaluwarsa · Abort"]
    F -->|Valid| G["c.Set context<br/>user_id · email · whatsapp_number"]

    G --> H["handler.GetProfile"]
    H --> I{"user_id exists<br/>in context?"}
    I -->|No| E4["401 · Identitas pengguna<br/>tidak ditemukan pada token"]
    I -->|Yes| J{"Type of user_id<br/>is float64 or int64 ?"}
    J -->|Other type| E5["401 · Format ID pengguna<br/>tidak valid"]
    J -->|Valid| K["service.GetUserProfile<br/>SELECT users WHERE id = ?"]
    K -->|Not found| E6["404 · Akun tidak ditemukan"]
    K -->|Found| L(["200 OK<br/>id · email · whatsapp_number<br/>created_at · updated_at"])
```

---

### 📎 Diagram Relasi Alur dengan Tabel Database

Menunjukkan tabel mana yang diakses pada setiap endpoint:

```mermaid
%%{init: {"theme":"default"}}%%
flowchart LR
    subgraph AUTH["/api/auth"]
        A1["1. register"]
        A2["2. login-email"]
        A3["3. wa/otp"]
        A4["4. wa/verify"]
        A5["5. forgot/otp"]
        A6["6. forgot/verify-otp"]
        A7["7. forgot/reset"]
        A8["8. forgot/reset-with-otp"]
    end
    A9["9. GET /user/profile"]

    AUTH --> U
    A9 --> U

    U[("USERS<br/>id PK · email UQ<br/>whatsapp_number UQ<br/>password_hash")]
    O[("OTP<br/>id PK · user_id FK<br/>kode_otp · tipe<br/>expired_at · is_used")]
    P[("PASSWORD_RESET<br/>id PK · user_id FK<br/>token<br/>expired_at · is_used")]

    A1 -->|"INSERT<br/>SELECT dup check"| U
    A2 -->|"SELECT + Bcrypt compare"| U
    A3 -->|"SELECT user,<br/>INSERT OTP"| U
    A3 -->|"INSERT<br/>tipe=LOGIN"| O
    A4 -->|"SELECT user,<br/>SELECT + MarkUsed OTP"| U
    A4 -->|"SELECT + UPDATE<br/>is_used=true"| O
    A5 -->|"SELECT user,<br/>INSERT OTP"| U
    A5 -->|"INSERT<br/>tipe=FORGOT_PASSWORD"| O
    A6 -->|"SELECT user,<br/>UPDATE OTP is_used"| U
    A6 -->|"SELECT + UPDATE OTP,<br/>INSERT reset token"| O
    A6 -->|"INSERT"| P
    A7 -->|"UPDATE password_hash<br/>+ updated_at"| U
    A7 -->|"SELECT + UPDATE<br/>is_used=true"| P
    A8 -->|"UPDATE password_hash"| U
    A8 -->|"SELECT + UPDATE OTP"| O
    A8 -->|"INSERT + UPDATE"| P
    A9 -->|"SELECT by id"| U

    O -.->|"FK user_id<br/>ON DELETE CASCADE"| U
    P -.->|"FK user_id<br/>ON DELETE CASCADE"| U
```

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
