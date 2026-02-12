# Phase 18: Internationalization (i18n)

> Implementasi dukungan multi-bahasa: Indonesia (default), English, Arabic (RTL).
> Dynamic language switch tanpa restart.

**Estimasi:** 3 hari
**Dependency:** Phase 06 (Mobile Shell)
**Output:** Seluruh UI tersedia dalam 3 bahasa, RTL support untuk Arabic.

---

## Task 18.1: i18n Backend Setup

**Input:** Existing API responses
**Output:** Localized error messages dan server-side strings

### Steps:
1. Buat `internal/i18n/i18n.go`:
   ```go
   type Language string

   const (
       LangID Language = "id" // Indonesian (default)
       LangEN Language = "en" // English
       LangAR Language = "ar" // Arabic
   )

   type Messages struct {
       Auth     AuthMessages
       Chat     ChatMessages
       Document DocumentMessages
       Error    ErrorMessages
   }

   var translations = map[Language]*Messages{
       LangID: {
           Auth: AuthMessages{
               OTPSent:          "Kode OTP telah dikirim",
               OTPInvalid:       "Kode OTP tidak valid",
               OTPExpired:       "Kode OTP telah kedaluwarsa",
               PhoneRequired:    "Nomor telepon wajib diisi",
               SessionExpired:   "Sesi telah berakhir, silakan login kembali",
           },
           Chat: ChatMessages{
               MessageDeleted:  "Pesan telah dihapus",
               GroupCreated:     "Grup telah dibuat",
               MemberAdded:     "%s telah ditambahkan",
               MemberRemoved:   "%s telah dikeluarkan",
               MemberLeft:      "%s keluar dari grup",
           },
           Document: DocumentMessages{
               Created:           "Dokumen dibuat",
               Locked:            "Dokumen dikunci",
               Unlocked:          "Dokumen dibuka",
               SignatureRequested: "Permintaan tanda tangan dikirim",
               Signed:            "Dokumen ditandatangani",
               CannotEditLocked:  "Dokumen terkunci, tidak dapat diedit",
           },
           Error: ErrorMessages{
               NotFound:       "Data tidak ditemukan",
               Unauthorized:   "Tidak memiliki akses",
               Forbidden:      "Akses ditolak",
               BadRequest:     "Permintaan tidak valid",
               InternalError:  "Terjadi kesalahan server",
               RateLimited:    "Terlalu banyak permintaan, coba lagi nanti",
           },
       },
       LangEN: { /* English translations */ },
       LangAR: { /* Arabic translations */ },
   }
   ```
2. Language detection from request:
   ```go
   // Middleware: extract language from Accept-Language header or user preference
   func LanguageMiddleware(next http.Handler) http.Handler {
       return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
           lang := r.Header.Get("Accept-Language")
           if lang == "" {
               lang = "id" // default
           }
           ctx := context.WithValue(r.Context(), "lang", Language(lang))
           next.ServeHTTP(w, r.WithContext(ctx))
       })
   }

   func GetLang(ctx context.Context) Language {
       if lang, ok := ctx.Value("lang").(Language); ok {
           return lang
       }
       return LangID
   }
   ```
3. User language preference stored in profile:
   ```go
   // Add to users table
   ALTER TABLE users ADD COLUMN language VARCHAR(2) DEFAULT 'id';
   ```

### Acceptance Criteria:
- [ ] 3 language files: ID, EN, AR
- [ ] Error messages localized
- [ ] Language from Accept-Language header
- [ ] User language preference stored
- [ ] Default: Indonesian

### Testing:
- [ ] Unit test: GetLang defaults to ID
- [ ] Unit test: middleware extracts language
- [ ] Unit test: translation lookup per language
- [ ] Unit test: format string with parameters (e.g., MemberAdded)

---

## Task 18.2: i18n Mobile Setup

**Input:** Phase 06 components
**Output:** React Native i18n infrastructure

### Steps:
1. Setup `i18next` + `react-i18next`:
   ```typescript
   // src/i18n/index.ts
   import i18n from 'i18next';
   import { initReactI18next } from 'react-i18next';
   import { getLocales } from 'react-native-localize';
   import AsyncStorage from '@react-native-async-storage/async-storage';

   import id from './locales/id.json';
   import en from './locales/en.json';
   import ar from './locales/ar.json';

   const LANGUAGE_KEY = 'user_language';

   i18n
     .use(initReactI18next)
     .init({
       resources: {
         id: { translation: id },
         en: { translation: en },
         ar: { translation: ar },
       },
       lng: 'id', // default
       fallbackLng: 'id',
       interpolation: { escapeValue: false },
     });

   // Restore saved language
   export async function initLanguage(): Promise<void> {
     const saved = await AsyncStorage.getItem(LANGUAGE_KEY);
     if (saved) {
       await i18n.changeLanguage(saved);
     } else {
       const deviceLang = getLocales()[0]?.languageCode;
       if (['id', 'en', 'ar'].includes(deviceLang)) {
         await i18n.changeLanguage(deviceLang);
       }
     }
   }

   export async function setLanguage(lang: 'id' | 'en' | 'ar'): Promise<void> {
     await i18n.changeLanguage(lang);
     await AsyncStorage.setItem(LANGUAGE_KEY, lang);
     // Also update API header
     api.defaults.headers['Accept-Language'] = lang;
   }

   export default i18n;
   ```
2. Buat locale files:
   ```json
   // src/i18n/locales/id.json
   {
     "common": {
       "save": "Simpan",
       "cancel": "Batal",
       "delete": "Hapus",
       "edit": "Edit",
       "search": "Cari",
       "loading": "Memuat...",
       "retry": "Coba Lagi",
       "done": "Selesai",
       "next": "Selanjutnya",
       "back": "Kembali",
       "close": "Tutup",
       "confirm": "Konfirmasi",
       "yes": "Ya",
       "no": "Tidak"
     },
     "auth": {
       "enterPhone": "Masukkan Nomor Telepon",
       "phoneLabel": "Nomor Telepon",
       "phonePlaceholder": "+62 812 3456 7890",
       "sendOTP": "Kirim Kode OTP",
       "enterOTP": "Masukkan Kode OTP",
       "otpSent": "Kode dikirim ke {{phone}}",
       "verify": "Verifikasi",
       "resendOTP": "Kirim Ulang",
       "resendIn": "Kirim ulang dalam {{seconds}}s",
       "setupProfile": "Atur Profil",
       "nameLabel": "Nama",
       "namePlaceholder": "Masukkan nama Anda"
     },
     "chat": {
       "title": "Chat",
       "newChat": "Chat Baru",
       "newGroup": "Grup Baru",
       "typeMessage": "Ketik pesan...",
       "online": "Online",
       "lastSeen": "Terakhir dilihat {{time}}",
       "typing": "sedang mengetik...",
       "you": "Anda",
       "messageDeleted": "Pesan ini telah dihapus",
       "noChats": "Belum Ada Chat",
       "startChat": "Mulai percakapan baru"
     },
     "group": {
       "createGroup": "Buat Grup",
       "groupName": "Nama Grup",
       "addMembers": "Tambah Anggota",
       "members": "Anggota",
       "admin": "Admin",
       "member": "Anggota",
       "leaveGroup": "Keluar Grup",
       "deleteGroup": "Hapus Grup",
       "groupInfo": "Info Grup"
     },
     "topic": {
       "topics": "Topik",
       "newTopic": "Topik Baru",
       "topicName": "Nama Topik",
       "noTopics": "Belum Ada Topik"
     },
     "document": {
       "title": "Dokumen",
       "newDocument": "Buat Dokumen",
       "selectTemplate": "Pilih Template",
       "noDocuments": "Belum Ada Dokumen",
       "createFirst": "Buat dokumen pertama untuk mulai berkolaborasi",
       "lock": "Kunci",
       "unlock": "Buka Kunci",
       "locked": "Terkunci",
       "draft": "Draft",
       "requestSignature": "Minta Tanda Tangan",
       "sign": "Tanda Tangani",
       "signed": "Ditandatangani",
       "pendingSignatures": "Menunggu TTD",
       "signaturesProgress": "{{signed}}/{{total}} tanda tangan",
       "enterPIN": "Masukkan PIN",
       "signConfirmation": "Dengan menandatangani, Anda menyetujui isi dokumen ini",
       "saving": "Menyimpan...",
       "saved": "Tersimpan",
       "history": "Riwayat"
     },
     "editor": {
       "paragraph": "Teks",
       "heading1": "Judul 1",
       "heading2": "Judul 2",
       "heading3": "Judul 3",
       "bulletList": "Daftar Bullet",
       "numberedList": "Daftar Nomor",
       "checklist": "Checklist",
       "table": "Tabel",
       "callout": "Callout",
       "code": "Kode",
       "toggle": "Toggle",
       "divider": "Pembatas",
       "quote": "Kutipan",
       "addBlock": "Tambah blok",
       "typeSlash": "Ketik / untuk menu"
     },
     "entity": {
       "entities": "Entity",
       "newEntity": "Buat Entity",
       "name": "Nama",
       "type": "Tipe",
       "fields": "Field",
       "addField": "Tambah Field",
       "linkedDocs": "Dokumen Terkait",
       "noEntities": "Belum Ada Entity",
       "tag": "Tag"
     },
     "search": {
       "title": "Cari",
       "placeholder": "Cari pesan, dokumen, kontak...",
       "all": "Semua",
       "messages": "Pesan",
       "documents": "Dokumen",
       "contacts": "Kontak",
       "noResults": "Tidak ditemukan hasil untuk '{{query}}'"
     },
     "settings": {
       "title": "Pengaturan",
       "profile": "Profil",
       "language": "Bahasa",
       "notifications": "Notifikasi",
       "storage": "Penyimpanan",
       "about": "Tentang",
       "logout": "Keluar",
       "logoutConfirm": "Yakin ingin keluar?"
     },
     "templates": {
       "empty": "Kosong",
       "meetingNotes": "Notulen Rapat",
       "shoppingList": "Daftar Belanja",
       "financialNotes": "Catatan Keuangan",
       "healthNotes": "Catatan Kesehatan",
       "agreement": "Kesepakatan Bersama",
       "farmingNotes": "Catatan Pertanian",
       "assetInventory": "Inventaris Aset"
     }
   }
   ```
3. Buat English and Arabic locale files with same structure
4. Arabic considerations:
   - RTL text direction
   - Mirrored UI layout

### Acceptance Criteria:
- [ ] 3 locale files complete (id, en, ar)
- [ ] All UI strings externalized
- [ ] Dynamic language switch without restart
- [ ] Language persisted to AsyncStorage
- [ ] Device language auto-detected on first launch
- [ ] API Accept-Language header updated

### Testing:
- [ ] Unit test: initLanguage from saved/device/default
- [ ] Unit test: setLanguage updates i18n + storage
- [ ] Component test: render with each language

---

## Task 18.3: RTL Support for Arabic

**Input:** Task 18.2
**Output:** Proper RTL layout for Arabic language

### Steps:
1. RTL detection and application:
   ```typescript
   // src/utils/rtl.ts
   import { I18nManager } from 'react-native';
   import RNRestart from 'react-native-restart';

   export function applyRTL(language: string): void {
     const isRTL = language === 'ar';

     if (I18nManager.isRTL !== isRTL) {
       I18nManager.allowRTL(isRTL);
       I18nManager.forceRTL(isRTL);
       // Restart required for RTL change
       RNRestart.restart();
     }
   }
   ```
2. RTL-aware style utilities:
   ```typescript
   // src/utils/styles.ts
   import { I18nManager, StyleSheet } from 'react-native';

   export const rtlStyle = (styles: any) => ({
     ...styles,
     ...(I18nManager.isRTL && {
       flexDirection: styles.flexDirection === 'row' ? 'row-reverse' : styles.flexDirection,
       textAlign: styles.textAlign === 'left' ? 'right' : styles.textAlign,
     }),
   });

   // Helper for directional margins/paddings
   export const marginStart = (value: number) =>
     I18nManager.isRTL ? { marginRight: value } : { marginLeft: value };

   export const marginEnd = (value: number) =>
     I18nManager.isRTL ? { marginLeft: value } : { marginRight: value };

   export const paddingStart = (value: number) =>
     I18nManager.isRTL ? { paddingRight: value } : { paddingLeft: value };
   ```
3. RTL-specific component adjustments:
   - Chat bubbles: sent (left for RTL), received (right for RTL)
   - Navigation: back arrow flipped
   - Icons: directional icons mirrored (arrows, chevrons)
   - Text alignment: auto (follows language direction)
   - Lists: item alignment flipped
4. Test screens:
   - Chat list in Arabic
   - Chat screen with Arabic messages
   - Document editor in Arabic
   - Settings screen in Arabic

### Acceptance Criteria:
- [ ] Arabic layout fully RTL
- [ ] Chat bubbles positioned correctly
- [ ] Navigation icons mirrored
- [ ] Text alignment follows direction
- [ ] Switch ID → AR requires restart (once)
- [ ] Style helpers work for RTL

### Testing:
- [ ] Unit test: rtlStyle transforms
- [ ] Unit test: margin/padding helpers
- [ ] Component test: chat screen in RTL
- [ ] Visual test: Arabic layout screenshots

---

## Task 18.4: Replace All Hardcoded Strings

**Input:** Task 18.2, all existing screens
**Output:** All screens use i18n strings

### Steps:
1. Audit all existing screens for hardcoded strings
2. Replace with `t()` calls:
   ```typescript
   // Before:
   <Text>Belum Ada Chat</Text>

   // After:
   import { useTranslation } from 'react-i18next';
   const { t } = useTranslation();
   <Text>{t('chat.noChats')}</Text>
   ```
3. Screens to update:
   - Auth screens (login, OTP, profile setup)
   - Chat list, chat screen
   - Group screens
   - Topic screens
   - Document screens
   - Entity screens
   - Search screen
   - Settings screen
   - All modals and sheets
4. Format dynamic strings:
   ```typescript
   // With interpolation:
   t('auth.otpSent', { phone: '+6281234567890' })
   // → "Kode dikirim ke +6281234567890"

   t('document.signaturesProgress', { signed: 2, total: 3 })
   // → "2/3 tanda tangan"
   ```

### Acceptance Criteria:
- [ ] Zero hardcoded strings in UI
- [ ] All screens render correctly in ID/EN/AR
- [ ] Dynamic strings with interpolation work
- [ ] Placeholder text localized
- [ ] Error messages localized

### Testing:
- [ ] Audit: grep for hardcoded Indonesian strings
- [ ] Component test: render screens in each language
- [ ] Visual test: screenshots in 3 languages

---

## Phase 18 Review

### Testing Checklist:
- [ ] Backend: localized error messages
- [ ] Frontend: 3 locale files complete
- [ ] Dynamic language switch works
- [ ] Arabic RTL layout correct
- [ ] All screens: no hardcoded strings
- [ ] Interpolation works for dynamic content
- [ ] Language persisted across app restart
- [ ] API Accept-Language header sent

### Review Checklist:
- [ ] i18n sesuai `spesifikasi-chatat.md` section 9.3
- [ ] Indonesian labels natural/colloquial
- [ ] Arabic: proper RTL, correct translations
- [ ] English: proper grammar/phrasing
- [ ] All 8 templates name translated
- [ ] Commit: `feat(i18n): implement multi-language support (ID/EN/AR)`
