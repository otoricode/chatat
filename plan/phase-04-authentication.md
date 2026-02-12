# Phase 04: Authentication System

> Implementasi SMS OTP dan Reverse OTP via WhatsApp untuk autentikasi.
> Phase ini menghasilkan auth flow lengkap sesuai spesifikasi.

**Estimasi:** 4 hari
**Dependency:** Phase 02 (Database), Phase 03 (API & WebSocket)
**Output:** Auth endpoints berfungsi, user bisa register dan login via OTP.

---

## Task 4.1: Phone Number Normalization

**Input:** API Foundation dari Phase 03
**Output:** Utility untuk normalisasi nomor HP ke format E.164

### Steps:
1. Buat `pkg/phone/normalize.go`:
   ```go
   // Normalize converts phone number to E.164 format
   // Input: "081234567890", "08-1234-567890", "+6281234567890"
   // Output: "+6281234567890"
   func Normalize(phone string, defaultCountry string) (string, error)

   // Validate checks if phone number is valid E.164
   func Validate(phone string) bool

   // Hash returns SHA-256 hash of phone number (for contact matching)
   func Hash(phone string) string
   ```
2. Install phone number library:
   ```bash
   go get github.com/nyaruka/phonenumbers
   ```
3. Implementasi:
   - Strip non-numeric characters (kecuali leading +)
   - Parse dengan phonenumbers library
   - Format ke E.164 (`+[country code][number]`)
   - Validate: must be valid mobile number
4. Hash function: SHA-256 untuk privacy saat contact matching

### Acceptance Criteria:
- [x] Indonesian numbers: `081xxx` → `+6281xxx`
- [x] International format pass-through: `+1xxx` → `+1xxx`
- [x] Invalid numbers return error
- [x] Hash consistent dan irreversible
- [x] Strip spaces, dashes, parentheses

### Testing:
- [x] Unit test: normalize Indonesian numbers (berbagai format)
- [x] Unit test: normalize international numbers
- [x] Unit test: invalid numbers
- [x] Unit test: hash consistency
- [x] Benchmark: normalize performance

---

## Task 4.2: OTP Service (SMS)

**Input:** Task 4.1, Redis dari Phase 02
**Output:** OTP generation, storage, dan verification

### Steps:
1. Buat `internal/service/otp_service.go`:
   ```go
   type OTPService interface {
       Generate(ctx context.Context, phone string) (string, error)
       Verify(ctx context.Context, phone string, code string) (bool, error)
       RateCheck(ctx context.Context, phone string) error
   }

   type otpService struct {
       redis  *redis.Client
       sms    SMSProvider
       config OTPConfig
   }

   type OTPConfig struct {
       Length     int           // 6
       TTL        time.Duration // 5 minutes
       MaxAttempts int          // 3
       CooldownBetween time.Duration // 60 seconds
       MaxPerDay  int          // 5
   }
   ```
2. Implementasi Generate:
   - Rate check: max 1 OTP per 60 detik per nomor
   - Rate check: max 5 OTP per hari per nomor
   - Generate 6-digit random code (crypto/rand)
   - Store di Redis: key `otp:{phone}`, value `{code, attempts}`, TTL 5 menit
   - Send via SMS provider
3. Implementasi Verify:
   - Get dari Redis
   - Compare code (constant-time)
   - Increment attempts counter
   - Max 3 attempts → invalidate
   - Jika berhasil → delete dari Redis
4. Buat SMS Provider interface:
   ```go
   type SMSProvider interface {
       Send(phone string, message string) error
   }
   ```
   - Implementasi placeholder/mock untuk development
   - Production: Twilio, Vonage, atau local provider

### Acceptance Criteria:
- [x] 6-digit OTP generated secara random
- [x] OTP expire setelah 5 menit
- [x] Max 3 attempts per OTP
- [x] Rate limit: 1 per 60 detik, 5 per hari
- [x] Constant-time comparison (timing attack prevention)
- [x] Redis storage dengan TTL

### Testing:
- [x] Unit test: generate OTP
- [x] Unit test: verify correct OTP
- [x] Unit test: verify wrong OTP
- [x] Unit test: OTP expiry
- [x] Unit test: max attempts exceeded
- [x] Unit test: rate limiting (cooldown)
- [x] Unit test: rate limiting (daily max)

---

## Task 4.3: Reverse OTP via WhatsApp

**Input:** Task 4.1, Redis dari Phase 02
**Output:** Reverse OTP flow menggunakan WhatsApp

### Steps:
1. Buat `internal/service/reverse_otp_service.go`:
   ```go
   type ReverseOTPService interface {
       InitSession(ctx context.Context, phone string) (*ReverseOTPSession, error)
       CheckVerification(ctx context.Context, sessionID string) (*VerificationResult, error)
   }

   type ReverseOTPSession struct {
       SessionID      string `json:"sessionId"`
       TargetWANumber string `json:"targetWANumber"` // nomor WA tujuan
       UniqueCode     string `json:"uniqueCode"`      // kode yang harus dikirim user
       ExpiresAt      time.Time `json:"expiresAt"`
   }
   ```
2. Implementasi InitSession:
   - Generate unique code (6 karakter alfanumerik)
   - Generate session ID
   - Store di Redis: key `reverse_otp:{sessionID}`, value `{phone, code, verified}`, TTL 5 menit
   - Return: nomor WA server + kode unik
3. Implementasi verification listener:
   - WhatsApp Business API webhook endpoint: `/api/v1/webhook/whatsapp`
   - Saat menerima pesan WA masuk:
     - Extract sender phone number
     - Extract message text (kode unik)
     - Match dengan session di Redis
     - Jika match → mark session as verified
4. Implementasi CheckVerification:
   - Polling endpoint for mobile app
   - Check Redis session status
   - Return verified/pending/expired
5. Buat WhatsApp provider interface:
   ```go
   type WhatsAppProvider interface {
       GetBusinessNumber() string
       OnMessageReceived(handler func(from string, body string))
   }
   ```
   - Placeholder untuk development (manual verification)
   - Production: WhatsApp Business API (Cloud API)

### Acceptance Criteria:
- [x] Session created dengan unique code
- [x] Server WhatsApp number returned to client
- [x] Incoming WA message matched to session
- [x] Session verified → user authenticated
- [x] Session expires after 5 minutes
- [x] Rate limiting same as SMS OTP

### Testing:
- [x] Unit test: init session
- [x] Unit test: verify via webhook
- [x] Unit test: check verification (pending, verified, expired)
- [x] Unit test: session expiry
- [x] Integration test: full reverse OTP flow (mock WA)

---

## Task 4.4: JWT Token Management

**Input:** Task 4.2 atau 4.3 selesai (auth verified)
**Output:** JWT token generation dan validation

### Steps:
1. Buat `internal/service/token_service.go`:
   ```go
   type TokenService interface {
       Generate(userID uuid.UUID) (*TokenPair, error)
       Validate(token string) (*Claims, error)
       Refresh(refreshToken string) (*TokenPair, error)
       Revoke(token string) error
   }

   type TokenPair struct {
       AccessToken  string `json:"accessToken"`
       RefreshToken string `json:"refreshToken"`
       ExpiresAt    int64  `json:"expiresAt"`
   }

   type Claims struct {
       UserID uuid.UUID `json:"userId"`
       jwt.RegisteredClaims
   }
   ```
2. Implementasi:
   - Access token: 15 menit expiry, signed with HS256
   - Refresh token: 30 hari expiry, stored in Redis
   - Revoke: add to Redis blacklist with TTL
3. Token storage di Redis:
   - Refresh token: key `refresh:{token}`, value `{userID}`, TTL 30 hari
   - Blacklist: key `blacklist:{token}`, value `1`, TTL = remaining token lifetime
4. Update auth middleware untuk check blacklist

### Acceptance Criteria:
- [x] Access token generated dengan proper claims
- [x] Refresh token stored di Redis
- [x] Token validation (signature, expiry, blacklist)
- [x] Token refresh flow berfungsi
- [x] Token revocation (logout)
- [x] Blacklisted token rejected

### Testing:
- [x] Unit test: generate token pair
- [x] Unit test: validate valid token
- [x] Unit test: validate expired token
- [x] Unit test: validate blacklisted token
- [x] Unit test: refresh token
- [x] Unit test: revoke token

---

## Task 4.5: Auth Handler & Endpoints

**Input:** Task 4.2, 4.3, 4.4 selesai
**Output:** Auth REST endpoints

### Steps:
1. Buat `internal/handler/auth_handler.go`:
   ```go
   type AuthHandler struct {
       otpService     service.OTPService
       reverseOTP     service.ReverseOTPService
       tokenService   service.TokenService
       userRepo       repository.UserRepository
   }
   ```
2. Endpoints:
   - `POST /api/v1/auth/otp/send`:
     - Body: `{"phone": "+6281234567890"}`
     - Normalize phone → rate check → generate OTP → send SMS
     - Response: `{"success": true, "data": {"expiresIn": 300}}`
   - `POST /api/v1/auth/otp/verify`:
     - Body: `{"phone": "+6281234567890", "code": "123456"}`
     - Verify OTP → create user if new → generate tokens
     - Response: `{"success": true, "data": {"accessToken": "...", "refreshToken": "...", "user": {...}, "isNewUser": true}}`
   - `POST /api/v1/auth/reverse-otp/init`:
     - Body: `{"phone": "+6281234567890"}`
     - Init session → return WA number + code
     - Response: `{"success": true, "data": {"sessionId": "...", "waNumber": "+62...", "code": "ABC123", "expiresIn": 300}}`
   - `POST /api/v1/auth/reverse-otp/check`:
     - Body: `{"sessionId": "..."}`
     - Check session status → if verified, create user + tokens
     - Response: same as otp/verify
   - `POST /api/v1/auth/refresh`:
     - Body: `{"refreshToken": "..."}`
     - Validate refresh → generate new pair
   - `POST /api/v1/auth/logout`:
     - Header: `Authorization: Bearer <token>`
     - Revoke access + refresh tokens
3. Create user logic:
   - If phone exists in DB → login (return existing user)
   - If phone not exists → create new user with placeholder name
   - `isNewUser: true` flag → mobile app shows profile setup screen

### Acceptance Criteria:
- [x] SMS OTP flow: send → verify → get tokens
- [x] Reverse OTP flow: init → user sends WA → check → get tokens
- [x] New user auto-created on first verification
- [x] Existing user: return user data with tokens
- [x] `isNewUser` flag accurate
- [x] Refresh token flow berfungsi
- [x] Logout revokes tokens
- [x] All error cases handled (invalid phone, wrong OTP, expired, rate limited)

### Testing:
- [x] Integration test: full SMS OTP flow
- [x] Integration test: full Reverse OTP flow
- [x] Integration test: new user registration
- [x] Integration test: existing user login
- [x] Integration test: token refresh
- [x] Integration test: logout
- [x] Integration test: error cases

---

## Task 4.6: Session Management (One Device)

**Input:** Task 4.4 selesai
**Output:** Satu nomor HP = satu device aktif

### Steps:
1. Buat device tracking di Redis:
   - Key: `device:{userID}`, value: `{deviceID, refreshToken}`
   - Saat login: store new device, invalidate previous
2. Update auth flow:
   - Saat generate tokens → store device info
   - Jika sudah ada device aktif → revoke previous tokens
   - Push notification ke device lama: "Anda login di perangkat lain"
3. Device ID: generated by mobile app, stored in MMKV
4. Saat token refresh: check if device ID matches

### Acceptance Criteria:
- [x] Login di device baru → device lama logout otomatis
- [x] Previous refresh token invalidated
- [x] Device ID tracking di Redis
- [x] Token refresh only valid for same device

### Testing:
- [x] Unit test: single device enforcement
- [x] Unit test: device switch revokes old session
- [x] Unit test: device ID mismatch on refresh

---

## Phase 04 Review

### Testing Checklist:
- [x] SMS OTP: send → receive code → verify → get tokens
- [x] Reverse OTP: init → send WA → check → get tokens
- [x] New user: create + profile setup flag
- [x] Existing user: login + return user data
- [x] Token refresh: new pair generated
- [x] Logout: tokens revoked
- [x] Single device: old device kicked
- [x] Rate limit: enforced on OTP endpoints
- [x] Invalid inputs: proper error responses
- [x] `go test ./...` — semua test pass

### Review Checklist:
- [x] Auth flow sesuai `spesifikasi-chatat.md` section 2
- [x] Phone normalization E.164
- [x] OTP security: crypto/rand, constant-time compare
- [x] JWT security: proper secret, appropriate TTL
- [x] Rate limiting effective
- [x] Error codes sesuai `docs/error-handling.md`
- [x] Commit: `feat(auth): implement OTP and Reverse OTP authentication`
