# Specs Checklist — Chatat

> Breakdown detail dari `spesifikasi-chatat.md` v3.0
> Setiap item dipecah menjadi checklist granular.
> Status: `[ ]` = belum tercoverage oleh plan, `[x]` = sudah tercoverage.
> Referensi phase ditandai dengan `→ Phase XX`.

---

## 1. Ikhtisar Aplikasi

### 1.1 Platform & Target
- [x] Mobile app React Native (iOS & Android) → Phase 01 (Task 1.2), Phase 06
- [x] Tidak ada versi web atau desktop → Phase 01 (mobile-only setup)
- [x] Koneksi antar pengguna berbasis nomor HP → Phase 04, 05
- [x] Pengguna yang saling simpan nomor HP & terdaftar → otomatis terkoneksi → Phase 05 (Task 5.2)

### 1.2 Arsitektur Komunikasi
- [x] Chat Personal (1-on-1) → Phase 07
- [x] Chat Grup (3+ orang) → Phase 08
- [x] Topik — ruang diskusi terfokus, lahir dari Chat Personal atau Grup → Phase 10
- [x] Dokumen bisa hidup di semua level (Chat Personal, Grup, Topik) → Phase 12 (Task 12.1)
- [x] Dokumen model hybrid: card inline di chat + tab Dokumen terpisah → Phase 14 (Task 14.4)

### 1.3 Fitur Utama (Ringkasan)
- [x] Chat WhatsApp-Style (personal & group) → Phase 07, 08
- [x] Kontak berbasis nomor HP → Phase 05
- [x] Topik (ruang diskusi terfokus) → Phase 10
- [x] Dokumen kolaboratif Notion-style → Phase 12, 13
- [x] Penguncian dokumen permanen → Phase 14

### 1.4 Spesifikasi Teknis
- [x] Nama aplikasi: Chatat → Phase 01
- [x] Versi: 3.0.0 → Phase 26 (build versioning)
- [x] Bahasa antarmuka: Indonesia, English, Arabic → Phase 18
- [x] Penyimpanan: Local-first + server relay → Phase 19

---

## 2. Sistem Autentikasi & Kontak

### 2.1 Metode Autentikasi
- [x] Verifikasi nomor HP via SMS OTP → Phase 04 (Task 4.2)
- [x] Verifikasi via Reverse OTP (WhatsApp) → Phase 04 (Task 4.3)
- [x] Tidak ada username, email, atau password → Phase 04 (phone-only design)

### 2.1.1 Alur SMS OTP
- [x] Input nomor HP format internasional (+62xxx, +1xxx) → Phase 04 (Task 4.1)
- [x] Sistem kirim OTP 6 digit via SMS → Phase 04 (Task 4.2)
- [x] Input kode OTP → verifikasi → Phase 04 (Task 4.2, 4.5)
- [x] Jika cocok, akun aktif → Phase 04 (Task 4.5)
- [x] Isi profil: Nama dan Avatar setelah verifikasi → Phase 05 (Task 5.1), Phase 06 (Task 6.4)

### 2.1.2 Alur Reverse OTP via WhatsApp
- [x] Input nomor HP → Phase 04 (Task 4.3)
- [x] Sistem tampilkan nomor WA tujuan + kode unik → Phase 04 (Task 4.3)
- [x] Pengguna kirim pesan WA berisi kode unik ke nomor tujuan → Phase 04 (Task 4.3)
- [x] Sistem deteksi pesan WA masuk & verifikasi nomor → Phase 04 (Task 4.3)
- [x] Nomor terverifikasi, akun aktif → Phase 04 (Task 4.3, 4.5)

### 2.2 Detail Autentikasi
- [x] Identitas: Nomor HP (unik per pengguna) → Phase 02 (Task 2.3), Phase 04
- [x] Sesi tersimpan di perangkat, tidak perlu login ulang → Phase 04 (Task 4.4)
- [x] Profil: Nama + Avatar (emoji) → Phase 05 (Task 5.1), Phase 06 (Task 6.4)
- [x] Multi-device: 1 nomor HP = 1 perangkat aktif → Phase 04 (Task 4.6)

### 2.3 Registrasi Pengguna Baru
- [x] Install & buka pertama kali → Phase 06 (Task 6.4)
- [x] Input nomor HP → Phase 04 (Task 4.1), Phase 06 (Task 6.4)
- [x] Verifikasi via SMS OTP atau Reverse OTP → Phase 04 (Tasks 4.2, 4.3)
- [x] Isi profil: Nama + pilih Avatar (emoji) → Phase 06 (Task 6.4)
- [x] Akun aktif → Phase 04 (Task 4.5)
- [x] Pengguna muncul di daftar kontak pengguna lain yg punya nomor HP-nya → Phase 05 (Task 5.2)

### 2.4 Sistem Kontak Berbasis Nomor HP
- [x] Koneksi berdasarkan nomor HP dari kontak ponsel → Phase 05 (Task 5.2)
- [x] Saat registrasi, minta akses kontak ponsel → Phase 07 (Task 7.6)
- [x] Cocokkan nomor HP dengan pengguna terdaftar → Phase 05 (Task 5.2, SHA-256 hash matching)
- [x] Kontak otomatis muncul jika nomor HP sudah terdaftar → Phase 05 (Task 5.2)
- [x] Bisa memulai chat personal dari daftar kontak → Phase 07 (Task 7.6)
- [x] Daftar kontak: avatar, nama, nomor HP, status online/offline → Phase 05 (Task 5.3, 5.4)
- [x] Bisa memulai chat dengan input nomor HP manual → Phase 05 (Task 5.3, search by phone)

### 2.5 Manajemen Sesi
- [x] Sesi login persisten di perangkat → Phase 04 (Task 4.4), Phase 06 (Task 6.5, MMKV)
- [x] Tidak perlu login ulang kecuali install ulang atau tekan "Keluar" → Phase 04, Phase 21 (Task 21.4)
- [x] 1 nomor HP = 1 perangkat aktif bersamaan → Phase 04 (Task 4.6)

### 2.6 Profil Pengguna
- [x] Field: id (String unik) → Phase 02 (Task 2.3, UUID)
- [x] Field: name (Teks) → Phase 02 (Task 2.3)
- [x] Field: phone (String, format internasional) → Phase 02 (Task 2.3)
- [x] Field: avatar (Emoji) → Phase 02 (Task 2.3)
- [x] Field: status (String, opsional) → Phase 02 (Task 2.3)
- [x] Field: lastSeen (ISO DateTime) → Phase 02 (Task 2.3)

---

## 3. Fitur Chat (WhatsApp-Style)

### 3.1 Jenis Chat
- [x] Chat Personal (1-on-1) → Phase 07
- [x] Chat Grup (beberapa pengguna) → Phase 08

### 3.2 Chat Personal (1-on-1)
- [x] Pilih kontak dari daftar kontak untuk memulai → Phase 07 (Task 7.6)
- [x] Satu ruang chat persisten per pasangan pengguna → Phase 07 (Task 7.1, GetOrCreatePersonalChat)
- [x] Header: avatar, nama kontak, status online/terakhir dilihat → Phase 07 (Task 7.5), Phase 09 (Task 9.5)
- [x] Tab Chat (💬): percakapan teks + dokumen inline → Phase 07 (Task 7.5), Phase 14 (Task 14.4)
- [x] Tab Dokumen (📄): daftar semua dokumen chat personal → Phase 14 (Task 14.4)
- [x] Bisa membuat Topik dari chat personal → Phase 10 (Task 10.4)

### 3.3 Chat Grup
- [x] Buat grup baru dengan pilih beberapa kontak → Phase 08 (Task 8.1, 8.3)
- [x] Grup: nama grup, ikon/foto grup (emoji), daftar anggota → Phase 08 (Task 8.1)
- [x] Admin grup (pembuat) bisa menambah/mengeluarkan anggota → Phase 08 (Task 8.1)
- [x] Semua anggota bisa kirim pesan → Phase 08 (Task 8.1)
- [x] Tab Chat (💬): percakapan teks + dokumen inline → Phase 08 (Task 8.4), Phase 14 (Task 14.4)
- [x] Tab Dokumen (📄): daftar semua dokumen grup → Phase 08 (Task 8.4), Phase 14 (Task 14.4)
- [x] Tab Topik (📌): daftar semua topik dalam konteks grup → Phase 08 (Task 8.4), Phase 10 (Task 10.4)
- [x] Bisa membuat Topik dari grup (sebagian/seluruh anggota) → Phase 10 (Task 10.4)

### 3.3.1 Membuat Grup Baru
- [x] Field: Nama Grup (teks, wajib) → Phase 08 (Task 8.3)
- [x] Field: Ikon Grup (emoji, wajib) → Phase 08 (Task 8.3)
- [x] Field: Anggota (multi-pilih kontak, min. 2, wajib) → Phase 08 (Task 8.1, 8.3)
- [x] Pembuat otomatis jadi admin → Phase 08 (Task 8.1)
- [x] Field: Deskripsi (teks, opsional) → Phase 08 (Task 8.3)

### 3.4 Fitur Chat Lengkap
- [x] Pesan teks (tanpa batas karakter) → Phase 07 (Task 7.2)
- [x] Bubble chat kiri/kanan (sendiri=hijau kanan, lain=abu kiri) → Phase 07 (Task 7.5)
- [x] Avatar & nama pengirim (di atas bubble dalam grup) → Phase 07 (Task 7.5), Phase 08 (Task 8.4)
- [x] Timestamp HH:MM di setiap bubble → Phase 07 (Task 7.5)
- [x] Pemisah tanggal otomatis saat hari berganti → Phase 07 (Task 7.5)
- [x] Status pesan: centang tunggal (terkirim), centang ganda (terbaca) → Phase 09 (Task 9.3)
- [x] Reply/Balas pesan (geser kanan, kutipan pesan asli) → Phase 07 (Task 7.2, 7.5)
- [x] Forward/Teruskan pesan ke chat/grup lain → Phase 07 (Task 7.2) ⚠️ *ditambahkan*
- [x] Hapus pesan (untuk diri sendiri atau semua) → Phase 07 (Task 7.2)
- [x] Auto-scroll ke pesan terbaru saat buka chat → Phase 07 (Task 7.5)
- [x] Kirim dengan Enter → Phase 07 (Task 7.5)
- [x] Tombol kirim ➤ di samping kolom input → Phase 07 (Task 7.5)
- [x] Terakhir dilihat ("terakhir dilihat pukul HH:MM") → Phase 09 (Task 9.5)
- [x] Typing indicator ("sedang mengetik...") → Phase 09 (Task 9.4)
- [x] Pratinjau di daftar chat (pesan terakhir + nama pengirim) → Phase 07 (Task 7.4)
- [x] Unread badge (jumlah pesan belum dibaca) → Phase 07 (Task 7.4)
- [x] Pencarian pesan dalam percakapan → Phase 17 (Task 17.3, in-chat search)
- [x] Panel emoji di keyboard → Phase 07 (Task 7.5, native keyboard emoji)
- [x] Kirim dokumen (card inline + masuk tab Dokumen) → Phase 14 (Task 14.4)

### 3.5 Daftar Chat (Chat List)
- [x] Halaman utama: daftar semua chat aktif → Phase 07 (Task 7.4)
- [x] Urut berdasarkan pesan terakhir (terbaru di atas) → Phase 07 (Task 7.1, 7.4)
- [x] Item: Avatar kontak/grup → Phase 07 (Task 7.4)
- [x] Item: Nama kontak/grup → Phase 07 (Task 7.4)
- [x] Item: Pratinjau pesan terakhir (dipotong) → Phase 07 (Task 7.4)
- [x] Item: Waktu pesan terakhir → Phase 07 (Task 7.4)
- [x] Item: Badge unread (jika ada) → Phase 07 (Task 7.4)
- [x] Item: Ikon pin untuk chat disematkan (opsional) → Phase 07 (Task 7.4)

### 3.6 Aksi Chat Tambahan
- [x] Pin chat (tekan lama → Pin, chat di atas daftar) → Phase 07 (Task 7.1, 7.4)
- [x] Arsipkan chat (tekan lama → Arsipkan, sembunyikan) → Phase 07 (Task 7.1, 7.4)
- [x] Baca semua (tekan lama → Tandai dibaca) → Phase 07 (Task 7.4) ⚠️ *ditambahkan*
- [x] Info grup (tekan nama grup di header → lihat anggota, nama, ikon) → Phase 08 (Task 8.5)

---

## 4. Fitur Topik (Ruang Diskusi)

### 4.1 Konsep Topik
- [x] Ruang diskusi terfokus, terpisah dari chat biasa → Phase 10
- [x] Selalu lahir dari konteks yang ada (Chat Personal / Grup) → Phase 10 (Task 10.1)
- [x] Setiap topik punya "rumah" (parent) yang jelas → Phase 02 (Task 2.5), Phase 10

### 4.2 Asal-Usul Topik (Parent)
- [x] Parent: Chat Personal → anggota otomatis kedua peserta → Phase 10 (Task 10.1, 10.4)
- [x] Parent: Chat Grup → anggota sebagian/seluruh member grup → Phase 10 (Task 10.1, 10.4)

### 4.3 Membuat Topik Baru
- [x] Dibuat dari dalam Chat Personal atau Grup → Phase 10 (Task 10.4)
- [x] Field: Ikon Topik (emoji, wajib) → Phase 10 (Task 10.4)
- [x] Field: Nama Topik (teks, wajib) → Phase 10 (Task 10.4)
- [x] Field: Anggota (dari anggota parent, wajib, min. 1) → Phase 10 (Task 10.1)
- [x] Dari personal: otomatis keduanya → Phase 10 (Task 10.1)
- [x] Dari grup: pilih sebagian/semua → Phase 10 (Task 10.4)
- [x] Field: Deskripsi (teks, opsional) → Phase 10 (Task 10.4)

### 4.3.1 Pilihan Ikon Topik
- [x] 10 ikon tersedia: 💬🏡🌾🏥📚💰🛒📋💼🤝 → Phase 10 (Task 10.4, emoji picker)

### 4.4 Fitur dalam Topik
- [x] Tab Diskusi (💬): ruang chat antar anggota + dokumen inline → Phase 10 (Task 10.4)
- [x] Tab Dokumen (📄): daftar semua dokumen topik → Phase 10 (Task 10.4)
- [x] Tombol 📄 di header topik → pintasan buat dokumen baru → Phase 10 (Task 10.4)

### 4.5 Aturan Keanggotaan Topik
- [x] Pembuat topik = admin, tidak bisa dikeluarkan → Phase 10 (Task 10.1)
- [x] Anggota harus berasal dari parent → Phase 02 (Task 2.5), Phase 10 (Task 10.1)
- [x] Topik dari personal: otomatis kedua peserta → Phase 10 (Task 10.1)
- [x] Topik dari grup: sebagian/seluruh member → Phase 10 (Task 10.1)
- [x] Admin bisa tambah anggota (dari parent) atau keluarkan anggota → Phase 10 (Task 10.1)

---

## 5. Fitur Dokumen Kolaboratif (Notion-Style)

### 5.1 Konsep Dokumen
- [x] Dokumen kolaboratif bergaya Notion → Phase 12, 13
- [x] Block-based editor → Phase 13 (Task 13.1)
- [x] Setiap elemen konten = block independen (tambah/hapus/pindah/format) → Phase 13 (Task 13.6)

### 5.2 Hybrid Model
- [x] Card inline di chat (preview, konteks temporal) → Phase 14 (Task 14.4)
- [x] Tab Dokumen (pengelolaan mudah, pencarian) → Phase 14 (Task 14.4)

### 5.2.1 Ownership per Konteks
- [x] Chat Personal → kedua peserta bisa akses → Phase 12 (Task 12.1)
- [x] Chat Grup → semua member grup bisa akses → Phase 12 (Task 12.1)
- [x] Topik (dari personal) → anggota topik → Phase 12 (Task 12.1)
- [x] Topik (dari grup) → anggota topik (subset member) → Phase 12 (Task 12.1)
- [x] Standalone → pemilik + kolaborator pilihan manual → Phase 12 (Task 12.1)

### 5.3 Tipe Block — Teks
- [x] Paragraf (teks biasa) → Phase 13 (Task 13.2)
- [x] Heading 1 (`# Judul`) → Phase 13 (Task 13.2)
- [x] Heading 2 (`## Sub-judul`) → Phase 13 (Task 13.2)
- [x] Heading 3 (`### Sub-sub-judul`) → Phase 13 (Task 13.2)
- [x] Bold (`**teks**`) → Phase 13 (Task 13.5)
- [x] Italic (`*teks*`) → Phase 13 (Task 13.5)
- [x] Strikethrough (`~~teks~~`) → Phase 13 (Task 13.5)
- [x] Inline Code (`` `kode` ``) → Phase 13 (Task 13.5) ⚠️ *ditambahkan ke toolbar*
- [x] Blockquote (`> kutipan`) → Phase 13 (Task 13.2)
- [x] Divider (`---`) → Phase 13 (Task 13.3)

### 5.3.2 Tipe Block — Daftar
- [x] Bullet List (`- item`) → Phase 13 (Task 13.2)
- [x] Numbered List (`1. item`) → Phase 13 (Task 13.2)
- [x] Checklist (`- [ ] item`, interaktif) → Phase 13 (Task 13.2)

### 5.3.3 Tipe Block — Data & Media
- [x] Tabel (ketik `/tabel`, kolom-baris dinamis) → Phase 13 (Task 13.3)
- [x] Callout (ketik `/callout`, kotak info + ikon emoji) → Phase 13 (Task 13.3)
- [x] Code Block (ketik `/kode`, syntax highlighting) → Phase 13 (Task 13.3)
- [x] Toggle (ketik `/toggle`, accordion buka/tutup) → Phase 13 (Task 13.3)

### 5.4 Slash Commands
- [x] Ketik `/` di baris kosong → menu pilihan block → Phase 13 (Task 13.4)
- [x] `/h1` atau `/heading1` → Heading 1 → Phase 13 (Task 13.4)
- [x] `/h2` atau `/heading2` → Heading 2 → Phase 13 (Task 13.4)
- [x] `/h3` atau `/heading3` → Heading 3 → Phase 13 (Task 13.4)
- [x] `/bullet` atau `/poin` → Bullet list → Phase 13 (Task 13.4)
- [x] `/angka` atau `/numbered` → Numbered list → Phase 13 (Task 13.4)
- [x] `/centang` atau `/checklist` → Checklist → Phase 13 (Task 13.4)
- [x] `/tabel` → Tabel baru → Phase 13 (Task 13.4)
- [x] `/callout` → Callout box → Phase 13 (Task 13.4)
- [x] `/kode` → Code block → Phase 13 (Task 13.4)
- [x] `/toggle` → Toggle/accordion → Phase 13 (Task 13.4)
- [x] `/pembatas` atau `/divider` → Divider → Phase 13 (Task 13.4)
- [x] `/kutipan` atau `/quote` → Blockquote → Phase 13 (Task 13.4)

### 5.5 Fitur Tabel Lanjutan
- [x] Tambah/hapus kolom (tombol `+` di kanan header) → Phase 13 (Task 13.3)
- [x] Tambah/hapus baris (tombol `+ Baris` di bawah) → Phase 13 (Task 13.3)
- [x] Resize kolom (drag pembatas kolom) → Phase 13 (Task 13.3) ⚠️ *ditambahkan*
- [x] Header row (baris pertama otomatis header style) → Phase 13 (Task 13.3)
- [x] Cell editing (klik untuk edit) → Phase 13 (Task 13.3)
- [x] Tipe kolom: Teks, Angka, Tanggal, Checkbox → Phase 13 (Task 13.3) ⚠️ *ditambahkan*

### 5.6 Toolbar Formatting
- [x] Floating toolbar saat seleksi teks → Phase 13 (Task 13.5)
- [x] Bold (Ctrl/Cmd+B) → Phase 13 (Task 13.5)
- [x] Italic (Ctrl/Cmd+I) → Phase 13 (Task 13.5)
- [x] Strikethrough (Ctrl/Cmd+Shift+S) → Phase 13 (Task 13.5)
- [x] Inline code (`<>`) → Phase 13 (Task 13.5) ⚠️ *ditambahkan*
- [x] Tambah link (🔗) → Phase 13 (Task 13.5)
- [x] Highlight warna → Phase 13 (Task 13.5) ⚠️ *ditambahkan*

### 5.7 Kolaborasi Dokumen
- [x] Pemilik: buat, edit, hapus, kunci, atur kolaborator → Phase 12 (Task 12.1, 12.3)
- [x] Editor: edit konten, tambah block, isi tabel → Phase 12 (Task 12.1)
- [x] Viewer: hanya bisa melihat → Phase 12 (Task 12.1)

### 5.8 Metadata Dokumen
- [x] Field: Judul (teks) → Phase 02 (Task 2.6), Phase 12
- [x] Field: Ikon (emoji) → Phase 02 (Task 2.6), Phase 12
- [x] Field: Cover (pilihan warna/gradien, opsional) → Phase 02 (Task 2.6, cover field)
- [x] Field: Label/Tag (multi-tag, kategorisasi) → Phase 02 (Task 2.6, document_tags), Phase 12 (Task 12.4)
- [x] Field: Kolaborator (multi-pilih kontak) → Phase 12 (Task 12.3)
- [x] Field: Konteks parent (auto-set: chat/grup/topik) → Phase 12 (Task 12.1)
- [x] Field: Entitas/Tag Subjek (multi-tag dinamis) → Phase 15

### 5.9 Entitas Dinamis (Entity Tags)
- [x] Label dinamis dibuat bebas pengguna → Phase 15 (Task 15.1)
- [x] Menandai subjek spesifik dalam dokumen → Phase 15 (Task 15.4)
- [x] Bisa berupa apa saja: lahan, kendaraan, anak, properti, proyek, hewan, perangkat → Phase 15 (Task 15.1)
- [x] Bisa berupa kontak dari daftar kontak → Phase 15 (Task 15.1, contact-to-entity)
- [x] Entitas kontak: link langsung ke profil Chatat → Phase 15 (Task 15.1)
- [x] Satu dokumen bisa punya beberapa entitas → Phase 02 (Task 2.7), Phase 15
- [x] Entitas global — bisa digunakan di dokumen mana pun → Phase 15 (Task 15.1)
- [x] Entitas sebagai filter di halaman Dokumen → Phase 15 (Task 15.4)
- [x] Ketik entitas baru (tersimpan otomatis) / pilih dari yang pernah dibuat / pilih dari kontak → Phase 15 (Task 15.4)
- [x] Tag kontak tidak otomatis beri akses (ikut konteks dokumen) → Phase 15 (Task 15.1)

### 5.10 Riwayat Dokumen
- [x] Log riwayat otomatis → Phase 14 (Task 14.5)
- [x] Aksi: Pembuatan ("Dibuat oleh [Nama]") → Phase 14 (Task 14.5)
- [x] Aksi: Pengeditan ("Diedit oleh [Nama]" + timestamp) → Phase 14 (Task 14.5)
- [x] Aksi: Kolaborator ditambah ("[Nama] ditambahkan sebagai [peran]") → Phase 14 (Task 14.5)
- [x] Aksi: Tanda tangan ("[Nama] menandatangani dokumen") → Phase 14 (Task 14.5)
- [x] Aksi: Penguncian ("Dokumen dikunci — semua tanda tangan terkumpul") → Phase 14 (Task 14.5)

### 5.11 Template Dokumen
- [x] Template: Kosong → Phase 12 (Task 12.3)
- [x] Template: Notulen Rapat (Agenda, Peserta, Pembahasan, Keputusan) → Phase 12 (Task 12.3)
- [x] Template: Daftar Belanja (Tabel: Barang, Jumlah, Harga, Total) → Phase 12 (Task 12.3)
- [x] Template: Catatan Keuangan (Tabel: Tanggal, Keterangan, Pemasukan, Pengeluaran, Saldo) → Phase 12 (Task 12.3)
- [x] Template: Catatan Kesehatan (Heading: Keluhan, Diagnosis, Obat, Dokter, Kunjungan) → Phase 12 (Task 12.3)
- [x] Template: Kesepakatan Bersama (Heading: Pihak, Isi, Ketentuan, Tanda Tangan) → Phase 12 (Task 12.3)
- [x] Template: Catatan Pertanian (Tabel: Lahan, Tanaman, Tanam, Panen, Catatan) → Phase 12 (Task 12.3)
- [x] Template: Inventaris Aset (Tabel: Aset, Jenis, Lokasi, Kondisi, Catatan) → Phase 12 (Task 12.3)

---

## 6. Penguncian Dokumen

### 6.1 Konsep Penguncian
- [x] Dokumen final yang tidak bisa diubah setelah dikunci → Phase 14 (Task 14.2)
- [x] Dua mekanisme: manual & tanda tangan digital → Phase 14 (Task 14.2)

### 6.2 Penguncian Manual
- [x] Pemilik bisa kunci kapan saja tanpa tanda tangan → Phase 14 (Task 14.2)
- [x] Alur: buka dokumen → menu ⋮ → "Kunci Dokumen" → konfirmasi → terkunci → Phase 14 (Task 14.3)
- [x] Dokumen terkunci permanen, tidak bisa diedit → Phase 14 (Task 14.2)

### 6.3 Penguncian dengan Tanda Tangan Digital
- [x] Pemilik aktifkan "Butuh tanda tangan" (status: Draft) → Phase 14 (Task 14.2)
- [x] Pemilik pilih penandatangan dari kontak (status: Menunggu Tanda Tangan) → Phase 14 (Task 14.3)
- [x] Simpan dokumen → badge ✍️ muncul → Phase 14 (Task 14.3)
- [x] Penandatangan buka & review → tekan "Tandatangani Sekarang" → Phase 14 (Task 14.3)
- [x] Setiap penandatangan menambah progres (1 dari N) → Phase 14 (Task 14.3)
- [x] Otomatis terkunci saat semua sudah tandatangan (🔒 TERKUNCI PERMANEN) → Phase 14 (Task 14.2)

### 6.4 Tampilan Status Tanda Tangan
- [x] ⏳ Menunggu (abu-abu) — belum tandatangan → Phase 14 (Task 14.3)
- [x] ✅ Ditandatangani · [Tanggal] (hijau) — sudah tandatangan + timestamp → Phase 14 (Task 14.3)
- [x] Banner hijau "Dokumen Terkunci" — semua selesai → Phase 14 (Task 14.3)

### 6.5 Badge Visual pada Kartu Dokumen
- [x] ✍️ Menunggu Tanda Tangan (ungu) — ada yang belum terkumpul → Phase 14 (Task 14.3, 14.4)
- [x] 🔒 Terkunci (kuning) — sudah dikunci → Phase 14 (Task 14.3, 14.4)
- [x] 📄 Draft (abu) — belum dikunci → Phase 14 (Task 14.3, 14.4)

### 6.6 Aturan Penguncian
- [x] Setelah terkunci, tidak ada yang bisa edit (termasuk pemilik) → Phase 14 (Task 14.2)
- [x] Terkunci tetap bisa dilihat semua kolaborator → Phase 14 (Task 14.2)
- [x] Penguncian permanen — tidak bisa dibuka → Phase 14 (Task 14.2)
- [x] Riwayat penguncian tercatat di log → Phase 14 (Task 14.5)
- [x] Belum dikunci → bisa dihapus pemilik; sudah dikunci → tidak bisa dihapus → Phase 14 (Task 14.2)

---

## 7. Navigasi & Antarmuka

### 7.1 Bottom Navigation
- [x] Dua tab: Chat (💬) dan Dokumen (📄) → Phase 06 (Task 6.1)
- [x] Tab Chat: daftar semua chat personal & grup → Phase 07 (Task 7.4)
- [x] Tab Dokumen: daftar semua dokumen lintas konteks → Phase 06 (Task 6.1), Phase 12 (Task 12.4)

### 7.2 Header & Aksi Cepat
- [x] Kiri: Logo/nama "Chatat" → Phase 06 (Task 6.3)
- [x] Kanan: Ikon pencarian 🔍, ikon profil (avatar) → Phase 06 (Task 6.3)

### 7.3 FAB (Floating Action Button)
- [x] Tombol (+) bulat hijau, pojok kanan bawah → Phase 06 (Task 6.3)
- [x] Tab Chat aktif → buka daftar kontak / buat grup baru → Phase 07 (Task 7.6)
- [x] Tab Dokumen aktif → buat dokumen standalone (pilih template/kosong) → Phase 12 (Task 12.1)

### 7.4 Halaman Kontak
- [x] Daftar pengguna terdaftar dari kontak ponsel → Phase 07 (Task 7.6)
- [x] Avatar, nama, status masing-masing → Phase 07 (Task 7.6)
- [x] Tap kontak → mulai/buka chat personal → Phase 07 (Task 7.6)
- [x] Tombol "Buat Grup" di atas daftar → Phase 07 (Task 7.6)
- [x] Field pencarian (nama atau nomor HP) → Phase 07 (Task 7.6)

### 7.5 Filter Dokumen
- [x] Baris filter horizontal di tab Dokumen → Phase 12 (Task 12.4)
- [x] Filter: Semua (🗂️) → Phase 12 (Task 12.4)
- [x] Filter: Dikunci (🔒) → Phase 12 (Task 12.4)
- [x] Filter: Menunggu Tanda Tangan (✍️) → Phase 12 (Task 12.4)
- [x] Filter: Draft (📄) → Phase 12 (Task 12.4)
- [x] Filter: Per Label (🏷️) → Phase 12 (Task 12.4)
- [x] Filter: Per Entitas (📌) → Phase 15 (Task 15.4)

### 7.6 Pencarian Global
- [x] Pencarian lintas fitur dari header → Phase 17 (Task 17.3)
- [x] Cari pesan dalam chat → Phase 17 (Task 17.1, 17.3)
- [x] Cari kontak (nama/nomor) → Phase 17 (Task 17.1, 17.3)
- [x] Cari dokumen (judul/konten) → Phase 17 (Task 17.1, 17.3)
- [x] Cari topik (nama) → Phase 17 (Task 17.1, 17.3)

---

## 8. Penyimpanan Data

### 8.1 Arsitektur Penyimpanan
- [x] Local-first dengan server relay dan sync → Phase 19
- [x] Pesan chat & topik: lokal (device), server hanya relay → Phase 19 (Task 19.1, 19.2)
- [x] Dokumen kolaboratif: server + sync ke kolaborator → Phase 14 (Task 14.1), Phase 19 (Task 19.3)
- [x] Profil & kontak: server + cache lokal → Phase 05, Phase 19
- [x] Entitas: server + cache lokal → Phase 15, Phase 19
- [x] Media (foto, file): upload ke server, download penerima, simpan lokal → Phase 11
- [x] Backup: cloud opsional (Google Drive Android / iCloud iOS) → Phase 20

### 8.1.1 Prinsip Utama
- [x] Store-and-forward: server simpan sementara, hapus setelah terkirim → Phase 19 (Task 19.2)
- [x] Lokal sebagai sumber utama (riwayat chat di perangkat) → Phase 19 (Task 19.1)
- [x] Status pengiriman: ✓ terkirim server, ✓✓ terkirim penerima, biru = dibaca → Phase 09 (Task 9.3)

### 8.2 Struktur Data

#### 8.2.1 User
- [x] id (String), name (String), phone (String), avatar (String/emoji) → Phase 02 (Task 2.3)
- [x] status (String), lastSeen (ISO DateTime), createdAt (ISO DateTime) → Phase 02 (Task 2.3)

#### 8.2.2 Chat
- [x] id (String), type (enum: personal/group) → Phase 02 (Task 2.4)
- [x] name (String|null), icon (String|null) → Phase 02 (Task 2.4)
- [x] memberIds (Array String), adminIds (Array String) → Phase 02 (Task 2.4, chat_members)
- [x] messages (Array: {id, senderId, text, replyTo, at, readBy}) → Phase 02 (Task 2.4, 2.8)
- [x] pinnedAt (ISO DateTime|null), createdAt (ISO DateTime) → Phase 02 (Task 2.4)

#### 8.2.3 Topic
- [x] id, name, icon (emoji), description → Phase 02 (Task 2.5)
- [x] parentType (enum: personal/group), parentId → Phase 02 (Task 2.5)
- [x] memberIds, adminIds → Phase 02 (Task 2.5, topic_members)
- [x] messages (Array: {id, senderId, text, replyTo, at}) → Phase 02 (Task 2.5, topic_messages)
- [x] createdAt → Phase 02 (Task 2.5)

#### 8.2.4 Document
- [x] id, title, icon (emoji), cover (String|null) → Phase 02 (Task 2.6)
- [x] blocks (Array Object), tags (Array String), entities (Array String) → Phase 02 (Task 2.6, 2.7)
- [x] ownerId, collaborators (Array: {userId, role}) → Phase 02 (Task 2.6)
- [x] topicId, chatId, groupId (konteks parent) → Phase 02 (Task 2.6)
- [x] requireSigs (Boolean), signerIds, sigs (Map: {userId: {at, name}}) → Phase 02 (Task 2.6)
- [x] locked (Boolean), lockedAt, lockedBy (manual/signatures) → Phase 02 (Task 2.6)
- [x] history (Array: [{at, action, userId}]) → Phase 02 (Task 2.6, document_history)
- [x] createdAt, updatedAt → Phase 02 (Task 2.6)

#### 8.2.5 Block
- [x] id (String), type (enum: 13 tipe) → Phase 02 (Task 2.6)
- [x] content (String|null), checked (Boolean|null, checklist) → Phase 02 (Task 2.6)
- [x] children (Array Block|null, toggle), rows (Array Array|null, table) → Phase 02 (Task 2.6)
- [x] columns (Array: {name, type}|null, table) → Phase 02 (Task 2.6)
- [x] language (String|null, code block) → Phase 02 (Task 2.6)
- [x] emoji (String|null, callout), color (String|null) → Phase 02 (Task 2.6)

---

## 9. Desain Visual & Antarmuka

### 9.1 Prinsip Desain
- [x] Mobile-First (max 430px lebar) → Phase 06 (Task 6.2)
- [x] Dark Theme (#0F1117 background) → Phase 06 (Task 6.2)
- [x] WhatsApp-Familiar (layout & pattern interaksi) → Phase 06 (Task 6.2), Phase 07
- [x] Kontras Tinggi (teks putih, aksen hijau) → Phase 06 (Task 6.2)
- [x] Minimal Kognitif (satu tugas per layar) → Phase 06 (Task 6.1)
- [x] Umpan Balik Visual (animasi, typing, status pesan) → Phase 09

### 9.2 Palet Warna
- [x] Background: #0F1117 → Phase 06 (Task 6.2)
- [x] Surface: #1A1D27 → Phase 06 (Task 6.2)
- [x] Surface 2: #222637 → Phase 06 (Task 6.2)
- [x] Border: #2E3348 → Phase 06 (Task 6.2)
- [x] Teks Utama: #E8EAF0 → Phase 06 (Task 6.2)
- [x] Teks Muted: #6B7280 → Phase 06 (Task 6.2)
- [x] Aksen Hijau: #6EE7B7 (CTA, bubble sendiri, online) → Phase 06 (Task 6.2)
- [x] Aksen Ungu: #818CF8 (nama pengirim grup, badge TTD) → Phase 06 (Task 6.2)
- [x] Bahaya: #F87171 (error, logout, hapus) → Phase 06 (Task 6.2)
- [x] Peringatan: #FBBF24 (badge terkunci) → Phase 06 (Task 6.2)
- [x] Aksen Biru: #60A5FA (link, judul dokumen) → Phase 06 (Task 6.2)

### 9.3 Tipografi
- [x] Plus Jakarta Sans (font UI) → Phase 06 (Task 6.2)
- [x] Inter (font dokumen) → Phase 06 (Task 6.2)
- [x] JetBrains Mono (font kode) → Phase 06 (Task 6.2)

---

## 10. Roadmap Masa Depan (v1.1+)

> Items di section ini adalah fitur **masa depan** yang secara eksplisit ditandai
> di spesifikasi sebagai "Roadmap Pengembangan Masa Depan". Beberapa sudah
> diprioritaskan masuk v1.0, sisanya tetap di backlog.

### 10.1 Prioritas Tinggi
- [x] Foto & Media (kirim di chat, embed di dokumen) → Phase 11 *(diprioritaskan masuk v1.0)*
- [x] Notifikasi Push → Phase 16 *(diprioritaskan masuk v1.0)*
- [ ] Voice Message → *backlog v1.1+*
- [ ] Ekspor PDF → *backlog v1.1+*
- [x] Real-time Sync → Phase 09, 14 *(diprioritaskan masuk v1.0)*

### 10.2 Prioritas Menengah
- [ ] Mention (@nama) → *backlog v1.1+*
- [ ] Reaction emoji → *backlog v1.1+*
- [ ] Version history dokumen → *backlog v1.1+* (Phase 14 punya activity log, bukan full version history)
- [ ] Comment pada block → *backlog v1.1+*
- [ ] Kalender → *backlog v1.1+*
- [ ] Pengingat/alarm → *backlog v1.1+*

### 10.3 Jangka Panjang
- [ ] Panggilan suara/video → *backlog v2.0+*
- [ ] Workspace/Organisasi → *backlog v2.0+*
- [ ] Database Notion-style → *backlog v2.0+*
- [x] Backup & restore GDrive/iCloud → Phase 20 *(diprioritaskan masuk v1.0)*
- [ ] Enkripsi end-to-end → *backlog v2.0+*
- [ ] Bot/reminder → *backlog v2.0+*
- [ ] Widget home screen → *backlog v2.0+*

---

## 11. Glosarium
- [x] Definisi semua istilah penting didokumentasikan → spesifikasi-chatat.md Section 11

---

## Ringkasan Coverage

| Section | Total Items | Covered | Status |
|---------|------------|---------|--------|
| 1. Ikhtisar | 14 | 14 | ✅ 100% |
| 2. Autentikasi & Kontak | 33 | 33 | ✅ 100% |
| 3. Chat | 44 | 44 | ✅ 100% |
| 4. Topik | 20 | 20 | ✅ 100% |
| 5. Dokumen | 64 | 64 | ✅ 100% |
| 6. Penguncian | 19 | 19 | ✅ 100% |
| 7. Navigasi | 19 | 19 | ✅ 100% |
| 8. Penyimpanan | 17 | 17 | ✅ 100% |
| 9. Desain Visual | 17 | 17 | ✅ 100% |
| 10. Roadmap (future) | 16 | 4+12 backlog | ⏳ 4 diprioritaskan, 12 backlog |
| 11. Glosarium | 1 | 1 | ✅ 100% |
| **TOTAL (v1.0 scope)** | **248** | **248** | **✅ 100%** |

## Gap yang Ditemukan & Ditutup

Berikut item yang awalnya belum tercoverage dan sudah **ditambahkan ke phase terkait**:

| Gap | Ditambahkan ke | Keterangan |
|-----|---------------|------------|
| Forward/Teruskan pesan | Phase 07 (Task 7.2) | Forward message service + UI di long-press menu |
| Toolbar: Inline code | Phase 13 (Task 13.5) | Tombol `<>` di floating toolbar |
| Toolbar: Highlight warna | Phase 13 (Task 13.5) | Color highlight option di toolbar |
| Tabel: Resize kolom | Phase 13 (Task 13.3) | Drag-to-resize pada pembatas kolom |
| Tabel: Tipe kolom | Phase 13 (Task 13.3) | Column type selector saat buat kolom |
| Tandai dibaca (batch) | Phase 07 (Task 7.4) | Long-press action "Tandai dibaca" |

---

*Checklist ini di-generate dari `spesifikasi-chatat.md` v3.0 dan di-cross-reference dengan `plan/phase-01` s/d `plan/phase-27`.*
