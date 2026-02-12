# Phase 22: Security & Privacy

> Implementasi keamanan aplikasi: rate limiting, input validation,
> privacy controls, dan data protection.

**Estimasi:** 4 hari
**Dependency:** Phase 03 (API), Phase 04 (Auth)
**Output:** Hardened API, privacy controls, security best practices.

---

## Task 22.1: Rate Limiting & Abuse Prevention

**Input:** Phase 03 middleware, Redis
**Output:** Rate limiter per endpoint dan abuse detection

### Steps:
1. Buat `internal/middleware/rate_limiter.go`:
   ```go
   type RateLimiterConfig struct {
       Global       RateLimit // all endpoints
       Auth         RateLimit // OTP requests
       Message      RateLimit // message sending
       Upload       RateLimit // file uploads
       Search       RateLimit // search queries
   }

   type RateLimit struct {
       Requests int           // max requests
       Window   time.Duration // time window
   }

   var DefaultConfig = RateLimiterConfig{
       Global:  RateLimit{Requests: 100, Window: time.Minute},
       Auth:    RateLimit{Requests: 5, Window: 5 * time.Minute},
       Message: RateLimit{Requests: 60, Window: time.Minute},
       Upload:  RateLimit{Requests: 10, Window: time.Minute},
       Search:  RateLimit{Requests: 20, Window: time.Minute},
   }

   func RateLimitMiddleware(redis *redis.Client, config RateLimit, keyPrefix string) func(http.Handler) http.Handler {
       return func(next http.Handler) http.Handler {
           return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
               // Key: prefix + user_id (or IP for auth endpoints)
               key := keyPrefix + ":" + getUserID(r)
               count, err := redis.Incr(r.Context(), key).Result()
               if err != nil {
                   next.ServeHTTP(w, r)
                   return
               }

               if count == 1 {
                   redis.Expire(r.Context(), key, config.Window)
               }

               // Set headers
               w.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.Requests))
               w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(max(0, config.Requests-int(count))))

               if int(count) > config.Requests {
                   ttl, _ := redis.TTL(r.Context(), key).Result()
                   w.Header().Set("Retry-After", strconv.Itoa(int(ttl.Seconds())))

                   http.Error(w, `{"error":"rate_limited","message":"Terlalu banyak permintaan"}`, http.StatusTooManyRequests)
                   return
               }

               next.ServeHTTP(w, r)
           })
       }
   }
   ```
2. Apply rate limits:
   - Global: 100 req/min per user
   - Auth (OTP): 5 req/5 min per phone
   - Messages: 60 msg/min per user
   - Uploads: 10 uploads/min per user
   - Search: 20 queries/min per user
3. OTP brute-force protection:
   ```go
   // After 5 failed OTP attempts → lock for 30 minutes
   func (s *AuthService) VerifyOTP(ctx context.Context, phone, code string) error {
       key := "otp_attempts:" + phone
       attempts, _ := s.redis.Incr(ctx, key).Result()
       if attempts == 1 {
           s.redis.Expire(ctx, key, 30*time.Minute)
       }
       if attempts > 5 {
           return ErrTooManyAttempts
       }
       // verify...
       if success {
           s.redis.Del(ctx, key) // reset on success
       }
       return nil
   }
   ```
4. WebSocket rate limiting:
   - Max 30 messages/min per WebSocket connection
   - Disconnect abusive connections

### Acceptance Criteria:
- [ ] Rate limit headers in all responses
- [ ] 429 status on exceeded limits
- [ ] OTP brute-force protection (5 attempts → 30min lock)
- [ ] WebSocket message rate limiting
- [ ] Rate limits configurable per endpoint
- [ ] Redis-based counters (distributed)

### Testing:
- [ ] Unit test: rate limiter increments and blocks
- [ ] Unit test: OTP lockout after 5 attempts
- [ ] Unit test: rate limit headers set correctly
- [ ] Integration test: exceed limit → 429 response
- [ ] Integration test: WebSocket rate limiting

---

## Task 22.2: Input Validation & Sanitization

**Input:** All handlers from previous phases
**Output:** Comprehensive input validation

### Steps:
1. Buat `internal/validator/validator.go`:
   ```go
   import "github.com/go-playground/validator/v10"

   var validate = validator.New()

   func init() {
       // Custom validators
       validate.RegisterValidation("phone", validatePhone)
       validate.RegisterValidation("safecontent", validateSafeContent)
       validate.RegisterValidation("nohtml", validateNoHTML)
   }

   func validatePhone(fl validator.FieldLevel) bool {
       phone := fl.Field().String()
       // E.164 format: +[country][number]
       matched, _ := regexp.MatchString(`^\+[1-9]\d{6,14}$`, phone)
       return matched
   }

   func validateSafeContent(fl validator.FieldLevel) bool {
       content := fl.Field().String()
       // No script tags, no SQL injection patterns
       dangerous := []string{"<script", "javascript:", "onload=", "onerror="}
       lower := strings.ToLower(content)
       for _, d := range dangerous {
           if strings.Contains(lower, d) {
               return false
           }
       }
       return true
   }

   func validateNoHTML(fl validator.FieldLevel) bool {
       content := fl.Field().String()
       stripped := bluemonday.StrictPolicy().Sanitize(content)
       return content == stripped
   }
   ```
2. Sanitize all text inputs:
   - Message content: strip HTML, max 10,000 chars
   - Document title: strip HTML, max 200 chars
   - User name: strip HTML, max 50 chars
   - Group name: strip HTML, max 100 chars
   - Block content: preserve markdown, strip HTML tags
3. File upload validation:
   - Max file size: 25 MB
   - Allowed image types: jpg, png, gif, webp
   - Allowed file types: pdf, doc, docx, xls, xlsx
   - MIME type verification (not just extension)
   - Virus scan integration point (future)
4. SQL injection prevention:
   - All queries use parameterized statements (pgx)
   - No string concatenation in queries
   - Audit all repository methods

### Acceptance Criteria:
- [ ] Phone validation (E.164 format)
- [ ] HTML stripping on all text inputs
- [ ] Content length limits enforced
- [ ] File type/size validation
- [ ] MIME type verification
- [ ] No SQL injection vectors
- [ ] Validation errors return clear messages

### Testing:
- [ ] Unit test: phone validation (valid/invalid)
- [ ] Unit test: HTML sanitization
- [ ] Unit test: content length limits
- [ ] Unit test: file type validation
- [ ] Unit test: MIME type checking
- [ ] Security audit: review all queries for injection

---

## Task 22.3: Privacy Controls

**Input:** Phase 05 (User), Phase 21 (Settings)
**Output:** User privacy settings

### Steps:
1. Privacy settings:
   ```go
   type PrivacySettings struct {
       LastSeenVisibility string `json:"lastSeenVisibility"` // everyone, contacts, nobody
       OnlineVisibility   string `json:"onlineVisibility"`   // everyone, contacts, nobody
       ReadReceipts       bool   `json:"readReceipts"`       // show/hide blue checks
       ProfilePhotoVisibility string `json:"profilePhotoVisibility"` // everyone, contacts
   }
   ```
2. Privacy enforcement in APIs:
   ```go
   // GetUserOnlineStatus:
   // If target user's onlineVisibility == "nobody" → return null
   // If "contacts" → check if requester is in target's contacts
   // If "everyone" → return status

   // Read receipts:
   // If user.readReceipts == false → don't send blue check status
   // Still show delivered (✓✓) but not read (blue)
   ```
3. Privacy settings screen:
   ```typescript
   const PrivacySettingsScreen: React.FC = () => {
     return (
       <ScrollView style={styles.container}>
         <SettingSection title={t('privacy.lastSeen')}>
           <RadioGroup
             options={[
               { label: t('privacy.everyone'), value: 'everyone' },
               { label: t('privacy.contacts'), value: 'contacts' },
               { label: t('privacy.nobody'), value: 'nobody' },
             ]}
             selected={lastSeen}
             onSelect={updateLastSeen}
           />
         </SettingSection>

         <SettingSection title={t('privacy.online')}>
           <RadioGroup
             options={[
               { label: t('privacy.everyone'), value: 'everyone' },
               { label: t('privacy.contacts'), value: 'contacts' },
               { label: t('privacy.nobody'), value: 'nobody' },
             ]}
             selected={online}
             onSelect={updateOnline}
           />
         </SettingSection>

         <SettingSection title={t('privacy.readReceipts')}>
           <SettingRow
             label={t('privacy.showReadReceipts')}
             type="switch"
             value={readReceipts}
             onToggle={updateReadReceipts}
           />
           <Text style={styles.hint}>
             {t('privacy.readReceiptsHint')}
           </Text>
         </SettingSection>

         <SettingSection title={t('privacy.profilePhoto')}>
           <RadioGroup
             options={[
               { label: t('privacy.everyone'), value: 'everyone' },
               { label: t('privacy.contacts'), value: 'contacts' },
             ]}
             selected={profilePhoto}
             onSelect={updateProfilePhoto}
           />
         </SettingSection>
       </ScrollView>
     );
   };
   ```

### Acceptance Criteria:
- [ ] Last seen visibility: everyone/contacts/nobody
- [ ] Online status visibility: everyone/contacts/nobody
- [ ] Read receipts toggle
- [ ] Profile photo visibility: everyone/contacts
- [ ] Privacy enforced server-side
- [ ] Settings synced to server

### Testing:
- [ ] Unit test: privacy enforcement (last seen)
- [ ] Unit test: privacy enforcement (online status)
- [ ] Unit test: read receipts disabled
- [ ] Component test: privacy settings screen
- [ ] Integration test: hidden last seen from non-contact

---

## Task 22.4: Data Protection & Security Headers

**Input:** Phase 03 (API)
**Output:** Security hardening

### Steps:
1. Security headers middleware:
   ```go
   func SecurityHeaders(next http.Handler) http.Handler {
       return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
           w.Header().Set("X-Content-Type-Options", "nosniff")
           w.Header().Set("X-Frame-Options", "DENY")
           w.Header().Set("X-XSS-Protection", "1; mode=block")
           w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
           w.Header().Set("Content-Security-Policy", "default-src 'none'")
           w.Header().Set("Cache-Control", "no-store")
           w.Header().Set("Pragma", "no-cache")
           next.ServeHTTP(w, r)
       })
   }
   ```
2. CORS configuration:
   ```go
   cors := cors.New(cors.Options{
       AllowedOrigins:   []string{}, // No web origins (mobile only)
       AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
       AllowedHeaders:   []string{"Authorization", "Content-Type", "Accept-Language"},
       MaxAge:           86400,
   })
   ```
3. JWT security:
   - Access token: 15 min TTL (short-lived)
   - Refresh token: 30 day TTL (rotating)
   - Token rotation: new refresh token on each use
   - Revocation: blacklist in Redis on logout
4. Sensitive data handling:
   - Never log passwords, tokens, or OTP codes
   - Mask phone numbers in logs (e.g., +62***7890)
   - PII scrubbing in error responses
   - Media URL signed with expiry
5. Mobile security:
   ```typescript
   // Secure storage for tokens
   import * as SecureStore from 'expo-secure-store';

   // Certificate pinning (optional, for high-security)
   // SSL pinning for API calls

   // Screenshot prevention (on sensitive screens)
   import { enableScreenCapture, disableScreenCapture } from 'expo-screen-capture';

   // Jailbreak/root detection
   import JailMonkey from 'jail-monkey';
   ```

### Acceptance Criteria:
- [ ] All security headers set
- [ ] CORS locked to mobile only
- [ ] JWT rotation on refresh
- [ ] Token revocation on logout
- [ ] No sensitive data in logs
- [ ] Phone masking in logs
- [ ] Secure token storage (mobile)

### Testing:
- [ ] Unit test: security headers present
- [ ] Unit test: JWT rotation generates new refresh token
- [ ] Unit test: revoked token rejected
- [ ] Unit test: phone masking
- [ ] Security audit: log file review

---

## Phase 22 Review

### Testing Checklist:
- [ ] Rate limiting: all endpoints
- [ ] OTP brute-force: lockout after 5 attempts
- [ ] Input validation: phone, content, files
- [ ] HTML sanitization
- [ ] Privacy controls: last seen, online, read receipts
- [ ] Security headers
- [ ] JWT rotation and revocation
- [ ] No sensitive data in logs
- [ ] `go test ./...` pass

### Review Checklist:
- [ ] Security sesuai `spesifikasi-chatat.md` section 9.4
- [ ] Rate limits appropriate for mobile usage
- [ ] Privacy options match WhatsApp patterns
- [ ] No OWASP Top 10 vulnerabilities
- [ ] Commit: `feat(security): implement security hardening and privacy controls`
