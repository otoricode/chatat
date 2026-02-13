# Integrasi WhatsApp (GOWA)

Chatat menggunakan [go-whatsapp-web-multidevice](https://github.com/aldinokemal/go-whatsapp-web-multidevice) (GOWA) sebagai WhatsApp gateway untuk fitur **Reverse OTP** dan pengiriman pesan WhatsApp.

## Arsitektur

```
User              GOWA Service           Chatat Server
  |                    |                       |
  | -- send WA msg --> |                       |
  |                    | -- webhook POST ----> |
  |                    |   /webhooks/whatsapp  |
  |                    |                       | (verify reverse OTP)
  |                    |                       |
  |                    | <-- POST /send/msg -- |
  |  <-- receive msg   |                       |
```

## Konfigurasi

### 1. Environment Variables

Tambahkan variabel berikut di file `.env` server:

```env
# WhatsApp (GOWA) - go-whatsapp-web-multidevice
WA_BASE_URL=http://localhost:3000
WA_WEBHOOK_SECRET=chatat-webhook-secret
WA_BUSINESS_PHONE=+628xxxxxxxxxx
```

| Variable | Deskripsi | Default |
|---|---|---|
| `WA_BASE_URL` | URL GOWA REST API | `http://localhost:3000` |
| `WA_WEBHOOK_SECRET` | HMAC secret untuk verifikasi webhook signature | `chatat-webhook-secret` |
| `WA_BUSINESS_PHONE` | Nomor WhatsApp bisnis (format E.164) | - |

### 2. Docker Compose

Service GOWA sudah dikonfigurasi di `docker-compose.yml`:

```yaml
whatsapp:
  image: aldinokemal2104/go-whatsapp-web-multidevice
  container_name: chatat-whatsapp
  restart: "on-failure"
  ports:
    - "3000:3000"
  volumes:
    - whatsapp_data:/app/storages
  environment:
    - APP_PORT=3000
    - APP_DEBUG=true
    - APP_OS=Chatat
    - APP_ACCOUNT_VALIDATION=false
    - WHATSAPP_WEBHOOK=http://host.docker.internal:8080/webhooks/whatsapp
    - WHATSAPP_WEBHOOK_SECRET=chatat-webhook-secret
    - WHATSAPP_WEBHOOK_EVENTS=message
    - WHATSAPP_AUTO_MARK_READ=true
  command:
    - rest
  extra_hosts:
    - "host.docker.internal:host-gateway"
```

**Penting:**
- `WHATSAPP_WEBHOOK` mengarah ke endpoint Chatat server. Jika server berjalan di host (bukan Docker), gunakan `host.docker.internal`.
- `WHATSAPP_WEBHOOK_SECRET` harus sama dengan `WA_WEBHOOK_SECRET` di server.
- `WHATSAPP_WEBHOOK_EVENTS=message` agar hanya event pesan yang dikirim ke webhook.

### 3. Menjalankan Service

```bash
# Start semua service (postgres, redis, whatsapp)
docker compose up -d

# Cek status
docker compose ps
```

### 4. Login WhatsApp

1. Buka browser: `http://localhost:3000`
2. Klik "Login" untuk mendapatkan QR code
3. Scan QR code dengan WhatsApp di HP
4. Setelah terhubung, catat nomor WhatsApp yang terdaftar
5. Set `WA_BUSINESS_PHONE` di `.env` server dengan nomor tersebut (format `+628xxxxxxxxxx`)

### 5. Verifikasi Konfigurasi

```bash
# Cek GOWA sudah berjalan
curl http://localhost:3000/app/status

# Response jika sudah login:
# {"code":"SUCCESS","message":"Connection status retrieved","results":{"is_connected":true,"is_logged_in":true}}

# Test kirim pesan (ganti nomor)
curl -X POST http://localhost:3000/send/message \
  -H "Content-Type: application/json" \
  -d '{"phone":"628xxxxxxxxxx@s.whatsapp.net","message":"Test dari Chatat"}'
```

## Cara Kerja Reverse OTP

1. User request reverse OTP via `POST /api/v1/auth/reverse-otp/init`
2. Server generate kode unik (misal: `A3X9K2`) dan return nomor WA bisnis
3. User mengirim pesan WA berisi kode tersebut ke nomor bisnis
4. GOWA menerima pesan dan mengirim webhook ke `POST /webhooks/whatsapp`
5. Server memverifikasi signature HMAC dan mencocokkan kode
6. User cek status via `POST /api/v1/auth/reverse-otp/check`

## Webhook Payload

GOWA mengirim webhook dengan format:

```json
{
  "event": "message",
  "device_id": "628xxx@s.whatsapp.net",
  "payload": {
    "id": "3EB0C127D7BACC83D6A1",
    "chat_id": "628xxx@s.whatsapp.net",
    "from": "628xxx@s.whatsapp.net",
    "from_name": "John Doe",
    "body": "A3X9K2",
    "is_from_me": false,
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```

Header `X-Hub-Signature-256` berisi HMAC SHA256 signature untuk verifikasi keamanan.

## Keamanan

- Webhook dilindungi dengan HMAC SHA256 signature verification
- Pesan dari diri sendiri (`is_from_me: true`) diabaikan
- Event non-message diabaikan
- Nomor pengirim dinormalisasi ke format E.164 sebelum diproses

## Troubleshooting

| Masalah | Solusi |
|---|---|
| GOWA tidak bisa connect | Buka `http://localhost:3000` dan login ulang via QR |
| Webhook tidak diterima | Pastikan `WHATSAPP_WEBHOOK` URL bisa diakses dari container |
| Signature verification gagal | Pastikan `WHATSAPP_WEBHOOK_SECRET` = `WA_WEBHOOK_SECRET` |
| Reverse OTP tidak terverifikasi | Cek log server untuk melihat apakah webhook diterima |

---

# Integrasi Firebase Cloud Messaging (FCM)

Chatat menggunakan Firebase Cloud Messaging untuk mengirim push notification ke perangkat mobile (Android & iOS).

## Arsitektur

```
Mobile App              Chatat Server              FCM
  |                          |                      |
  | -- register token -----> |                      |
  |   POST /notifications/   |                      |
  |        devices           |                      |
  |                          |                      |
  |                          | -- send notif ------> |
  |  <---- push notif -------|---------------------- |
  |                          |                      |
```

## Konfigurasi

### 1. Buat Firebase Project

1. Buka [Firebase Console](https://console.firebase.google.com/)
2. Klik "Add project" atau pilih project yang sudah ada
3. Aktifkan Cloud Messaging di menu **Project Settings > Cloud Messaging**

### 2. Generate Service Account Key

1. Di Firebase Console, buka **Project Settings > Service accounts**
2. Klik **Generate new private key**
3. Simpan file JSON yang didownload (contoh: `firebase-credentials.json`)
4. **JANGAN** commit file ini ke repository

### 3. Environment Variables

Tambahkan ke file `.env` server:

```env
# Firebase Cloud Messaging (Push Notifications)
FCM_CREDENTIALS_FILE=/path/to/firebase-credentials.json
```

Jika `FCM_CREDENTIALS_FILE` tidak diset atau kosong, server akan menggunakan **LogPushSender** yang hanya mencetak notifikasi ke log (untuk development).

### 4. Mobile Setup (Expo)

1. Buat project di [Expo Dashboard](https://expo.dev/)
2. Pastikan `projectId` di `app.json` sudah sesuai
3. Untuk Android: Upload Server Key dari Firebase ke Expo Dashboard
4. Untuk iOS: Upload APNs key ke Expo Dashboard

## Jenis Notifikasi

| Type | Trigger | Data |
|---|---|---|
| `message` | Pesan baru di chat personal | `chatId` |
| `group_message` | Pesan baru di chat group | `chatId` |
| `topic_message` | Pesan baru di topik | `topicId` |
| `signature_request` | Permintaan tanda tangan dokumen | `documentId` |
| `document_locked` | Dokumen dikunci | `documentId` |
| `group_invite` | Diundang ke grup | `chatId` |

## Mode Development

Tanpa `FCM_CREDENTIALS_FILE`, server menggunakan `LogPushSender`:

```
[PUSH] notification sent (log mode) token=ExponentPushToken... type=message title=Ahmad body=Ahmad: Halo
```

## Troubleshooting

| Masalah | Solusi |
|---|---|
| Push tidak terkirim | Cek apakah `FCM_CREDENTIALS_FILE` diset dan file ada |
| Token invalid | Token expired atau device unregistered, akan otomatis di-log |
| Permission denied di mobile | Pastikan user memberikan izin notifikasi |
| Token tidak terdaftar | Pastikan `POST /notifications/devices` dipanggil setelah login |
