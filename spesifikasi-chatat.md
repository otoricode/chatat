# 💬 Chatat
## Dokumen Spesifikasi Aplikasi

> **Versi 3.0 | Februari 2026**
> Mobile App (React Native) · WhatsApp-Style · Dokumen Kolaboratif

---

## Daftar Isi

1. [Ikhtisar Aplikasi](#1-ikhtisar-aplikasi)
2. [Sistem Autentikasi & Kontak](#2-sistem-autentikasi--kontak)
3. [Fitur Chat (WhatsApp-Style)](#3-fitur-chat-whatsapp-style)
4. [Fitur Topik (Ruang Diskusi)](#4-fitur-topik-ruang-diskusi)
5. [Fitur Dokumen Kolaboratif (Notion-Style)](#5-fitur-dokumen-kolaboratif-notion-style)
6. [Penguncian Dokumen](#6-penguncian-dokumen)
7. [Navigasi & Antarmuka](#7-navigasi--antarmuka)
8. [Penyimpanan Data](#8-penyimpanan-data)
9. [Desain Visual & Antarmuka](#9-desain-visual--antarmuka)
10. [Roadmap Pengembangan Masa Depan](#10-roadmap-pengembangan-masa-depan)
11. [Glosarium](#11-glosarium)

---

## 1. Ikhtisar Aplikasi

### 1.1 Deskripsi Umum

Chatat adalah aplikasi mobile yang menggabungkan pengalaman **chat seperti WhatsApp** dengan kemampuan **pembuatan dokumen kolaboratif seperti Notion**. Aplikasi ini dirancang untuk siapa saja — individu, keluarga, tim, komunitas, atau kelompok kerja — yang ingin berkomunikasi secara terorganisir sekaligus membuat, mengelola, dan mengunci dokumen bersama dalam satu platform terintegrasi.

Aplikasi ini hanya tersedia sebagai **mobile app** (iOS & Android) — seperti WhatsApp di awal peluncurannya. Tidak ada versi web atau desktop.

Koneksi antar pengguna berbasis **nomor HP** — persis seperti WhatsApp. Pengguna yang sudah menyimpan nomor HP satu sama lain dan sama-sama terdaftar di Chatat akan otomatis saling terkoneksi sebagai kontak.

### 1.2 Arsitektur Komunikasi

Chatat memiliki tiga lapisan komunikasi yang saling terhubung:

```
Chat Personal (1-on-1) ───┬─── bisa bikin Topik
                           │
Chat Grup (3+ orang) ────┤─── bisa bikin Topik
                           │
              Topik ──────┘─── diskusi terfokus
```

- **Chat Personal:** Percakapan langsung 1-on-1, seperti WA biasa.
- **Chat Grup:** Percakapan banyak orang (min. 3), seperti WA group.
- **Topik:** Ruang diskusi terfokus yang lahir dari Chat Personal atau Grup.

**Dokumen** bisa hidup di semua level — Chat Personal, Grup, atau Topik — dan menggunakan model **hybrid**: muncul sebagai card inline di chat sekaligus tersedia di tab Dokumen terpisah.

### 1.2 Tujuan & Manfaat

| Masalah Lama | Solusi Chatat |
|---|---|
| Komunikasi dan catatan tersebar di banyak aplikasi | Satu platform: chat + dokumen kolaboratif terintegrasi |
| Catatan/notulen tersebar di mana-mana | Dokumen kolaboratif Notion-style yang bisa diedit bersama |
| Tidak ada bukti kesepakatan yang tercatat formal | Penguncian dokumen permanen setelah semua pihak menyetujui |
| Diskusi topik tertentu tenggelam di group chat | Fitur Topik terpisah untuk diskusi terfokus per subjek |
| Catatan tidak terstruktur, sulit dibaca ulang | Editor dokumen dengan heading, tabel, checklist, dan format kaya |

### 1.3 Ringkasan Fitur Utama

| Fitur | Deskripsi Singkat |
|---|---|
| **Chat WhatsApp-Style** | Chat personal 1-on-1 dan group, persis seperti WhatsApp |
| **Kontak Berbasis Nomor HP** | Koneksi otomatis berdasarkan nomor HP, seperti WhatsApp |
| **Topik** | Ruang diskusi terfokus yang bisa dibuat dengan semua kontak |
| **Dokumen Kolaboratif** | Buat dokumen kaya format Notion-style bersama siapa saja |
| **Penguncian Dokumen** | Kunci dokumen permanen setelah semua pihak menandatangani |

### 1.4 Spesifikasi Teknis

| Atribut | Detail |
|---|---|
| **Nama Aplikasi** | Chatat |
| **Versi** | 3.0.0 |
| **Platform** | Mobile App (React Native) |
| **Target Perangkat** | Smartphone (iOS & Android) |
| **Kapasitas Pengguna** | Tidak terbatas (koneksi berbasis nomor HP) |
| **Bahasa Antarmuka** | Indonesia, English, Arabic |
| **Penyimpanan Data** | Local-first + server relay (WhatsApp-style) |

---

## 2. Sistem Autentikasi & Kontak

### 2.1 Metode Autentikasi (WhatsApp-Style)

Aplikasi menggunakan **verifikasi nomor HP via SMS OTP** atau **Reverse OTP** sebagai metode autentikasi — persis seperti WhatsApp. Tidak ada username, email, atau password.

#### 2.1.1 Alur Verifikasi SMS OTP

1. Pengguna membuka aplikasi untuk pertama kali.
2. Memasukkan nomor HP (format internasional: `+62xxx`, `+1xxx`, dll).
3. Sistem mengirim kode OTP 6 digit via SMS ke nomor tersebut.
4. Pengguna memasukkan kode OTP.
5. Jika cocok, nomor terverifikasi dan akun aktif.
6. Pengguna mengisi profil: Nama dan Avatar.

#### 2.1.2 Alur Reverse OTP via WhatsApp (Alternatif)

1. Pengguna memasukkan nomor HP.
2. Sistem menampilkan nomor WhatsApp tujuan dan kode unik.
3. Pengguna mengirim pesan WhatsApp berisi kode unik ke nomor tujuan tersebut.
4. Sistem mendeteksi pesan WhatsApp masuk dan memverifikasi nomor pengguna.
5. Nomor terverifikasi dan akun aktif.

> 💡 **Reverse OTP via WhatsApp** menghindari biaya SMS gateway karena pengguna yang mengirim pesan. Lebih murah dan reliable karena memanfaatkan WhatsApp yang sudah terinstall di hampir semua smartphone.

| Komponen | Detail |
|---|---|
| Identitas | Nomor HP (unik per pengguna) |
| Verifikasi | SMS OTP 6 digit atau Reverse OTP via WhatsApp |
| Sesi | Tersimpan di perangkat, tidak perlu login ulang |
| Profil | Nama + Avatar (diisi setelah verifikasi) |
| Multi-device | Satu nomor HP = satu perangkat aktif (seperti WA awal) |

### 2.2 Registrasi Pengguna Baru

1. Pengguna menginstall aplikasi dan membuka pertama kali.
2. Memasukkan nomor HP.
3. Verifikasi via SMS OTP atau Reverse OTP.
4. Mengisi profil: Nama dan memilih Avatar (emoji).
5. Akun aktif.
6. **Pengguna langsung muncul di daftar kontak pengguna lain yang memiliki nomor HP-nya di kontak ponsel.**

### 2.3 Sistem Kontak Berbasis Nomor HP (WhatsApp-Style)

Konsep kontak di Chatat mengadopsi pendekatan WhatsApp — berbasis nomor telepon:

- Koneksi antar pengguna terjadi berdasarkan **nomor HP dari daftar kontak ponsel**.
- Saat pengguna mendaftar, aplikasi meminta akses ke kontak ponsel dan mencocokkan nomor HP dengan pengguna lain yang sudah terdaftar di Chatat.
- Jika nomor HP seseorang yang ada di kontak ponsel sudah terdaftar di Chatat, orang tersebut otomatis muncul sebagai kontak di aplikasi.
- Pengguna bisa langsung memulai chat personal dengan siapa saja dari daftar kontak Chatat.
- Daftar kontak menampilkan: avatar, nama (dari kontak ponsel), nomor HP, dan status online/offline.
- Pengguna juga bisa memulai chat dengan memasukkan nomor HP secara manual.

> 💡 **Mengapa Berbasis Nomor HP?**
> Sama seperti WhatsApp — jika kamu sudah punya nomor HP seseorang dan dia sudah terdaftar di Chatat, kalian langsung bisa berkomunikasi. Tidak perlu menambahkan teman, mengirim undangan, atau menunggu persetujuan.

### 2.4 Manajemen Sesi

Sesi login disimpan secara persisten di perangkat. Pengguna tidak perlu login ulang kecuali menginstall ulang aplikasi atau menekan tombol **Keluar**. Satu nomor HP hanya bisa aktif di satu perangkat pada saat bersamaan (seperti WhatsApp awal).

### 2.5 Profil Pengguna

| Field | Tipe Data | Contoh |
|---|---|---|
| `id` | String unik | `u1`, `u2`, `u3` |
| `name` | Teks | Andi, Budi, Sari |
| `phone` | String | `+6281111111111` |
| `avatar` | Emoji | 👤 😊 🙋 |
| `status` | String | Teks status singkat (opsional) |
| `lastSeen` | ISO DateTime | Waktu terakhir aktif |

---

## 3. Fitur Chat (WhatsApp-Style)

### 3.1 Konsep Chat

Fitur chat di Chatat dirancang untuk **identik dengan pengalaman WhatsApp**. Terdapat dua jenis chat:

1. **Chat Personal (1-on-1):** Percakapan langsung antara dua pengguna.
2. **Chat Grup:** Percakapan dengan beberapa pengguna sekaligus.

### 3.2 Chat Personal (1-on-1)

- Pengguna memilih kontak dari daftar kontak untuk memulai chat.
- Setiap pasangan pengguna memiliki satu ruang chat yang persisten.
- Header chat menampilkan avatar, nama kontak, dan status online/terakhir dilihat.
- Chat personal memiliki **dua tab**:
  - **Tab Chat (💬):** Percakapan teks biasa + dokumen inline.
  - **Tab Dokumen (📄):** Daftar semua dokumen yang dimiliki chat personal ini.
- Dari chat personal, pengguna bisa **membuat Topik** untuk diskusi terfokus dengan orang yang sama.

> 💡 **Tidak perlu bikin grup cuma 2 orang.** Chat personal sudah menangani percakapan berdua. Jika butuh diskusi terfokus, buat Topik langsung dari chat personal.

### 3.3 Chat Grup

- Pengguna dapat membuat grup baru dengan memilih beberapa kontak.
- Grup memiliki: nama grup, ikon/foto grup (emoji), dan daftar anggota.
- Admin grup (pembuat) bisa menambah/mengeluarkan anggota.
- Semua anggota grup dapat mengirim pesan.
- Chat grup memiliki **tiga tab**:
  - **Tab Chat (💬):** Percakapan teks biasa + dokumen inline.
  - **Tab Dokumen (📄):** Daftar semua dokumen yang dimiliki grup ini.
  - **Tab Topik (📌):** Daftar semua topik yang dibuat dalam konteks grup ini.
- Dari grup, pengguna bisa **membuat Topik** dengan sebagian atau seluruh anggota grup.

#### 3.3.1 Membuat Grup Baru

| Field | Tipe | Wajib | Keterangan |
|---|---|---|---|
| Nama Grup | Teks bebas | Ya | Cth: "Tim Proyek", "Rencana Liburan", "Keluarga" |
| Ikon Grup | Emoji | Ya | Identifikasi visual |
| Anggota | Multi-pilih dari kontak | Ya (min. 2) | Pembuat otomatis menjadi admin |
| Deskripsi | Teks bebas | Tidak | Deskripsi singkat tujuan grup |

### 3.4 Fitur Chat Lengkap (WhatsApp-Style)

| Fitur | Deskripsi |
|---|---|
| **Pesan teks** | Kirim pesan teks bebas tanpa batas karakter |
| **Bubble chat kiri/kanan** | Pesan sendiri di kanan (hijau), pesan orang lain di kiri (abu) |
| **Avatar & nama pengirim** | Ditampilkan di atas bubble (dalam grup) |
| **Timestamp** | Waktu pesan dalam format HH:MM di setiap bubble |
| **Pemisah tanggal** | Garis pemisah otomatis saat hari berganti |
| **Status pesan** | Centang tunggal (terkirim), centang ganda (terbaca) |
| **Reply/Balas pesan** | Geser kanan pada pesan untuk membalas, menampilkan kutipan pesan asli |
| **Forward/Teruskan** | Teruskan pesan ke chat atau grup lain |
| **Hapus pesan** | Hapus pesan untuk diri sendiri atau untuk semua orang |
| **Auto-scroll** | Otomatis gulir ke pesan terbaru saat membuka chat |
| **Kirim dengan Enter** | Tekan Enter untuk mengirim pesan |
| **Tombol kirim** | Tombol ➤ di samping kolom input |
| **Terakhir dilihat** | Status "terakhir dilihat pukul HH:MM" di header chat |
| **Typing indicator** | Indikator "sedang mengetik..." saat lawan chat mengetik |
| **Pratinjau di daftar chat** | Pesan terakhir ditampilkan di daftar chat dengan nama pengirim |
| **Unread badge** | Badge jumlah pesan belum dibaca pada daftar chat |
| **Pencarian pesan** | Cari pesan dalam percakapan berdasarkan kata kunci |
| **Emoji** | Panel emoji di keyboard untuk mengirim emoji dalam pesan |
| **Kirim dokumen** | Buat atau lampirkan dokumen — muncul sebagai card inline di chat + otomatis masuk tab Dokumen |

### 3.5 Daftar Chat (Chat List)

Halaman utama aplikasi menampilkan daftar semua chat aktif, diurutkan berdasarkan pesan terakhir (terbaru di atas), persis seperti WhatsApp:

- Avatar kontak/grup
- Nama kontak/grup
- Pratinjau pesan terakhir (dipotong jika terlalu panjang)
- Waktu pesan terakhir
- Badge jumlah pesan belum dibaca (jika ada)
- Ikon pin untuk chat yang disematkan (opsional)

### 3.6 Aksi Chat Tambahan

| Aksi | Cara | Keterangan |
|---|---|---|
| Pin chat | Tekan lama → Pin | Chat disematkan di atas daftar |
| Arsipkan chat | Tekan lama → Arsipkan | Sembunyikan dari daftar utama |
| Baca semua | Tekan lama → Tandai dibaca | Hapus badge unread |
| Info grup | Tekan nama grup di header | Lihat anggota, nama, ikon grup |

---

## 4. Fitur Topik (Ruang Diskusi)

### 4.1 Konsep Topik

Topik adalah ruang diskusi **terfokus** yang terpisah dari chat biasa. Topik selalu **lahir dari konteks yang sudah ada** — bisa dari Chat Personal atau dari Grup. Ini memastikan setiap topik punya "rumah" (parent) yang jelas.

> 💡 **Topik vs Chat Biasa**
> - **Chat Personal/Grup:** Percakapan bebas, bisa membahas apa saja.
> - **Topik:** Diskusi terfokus dengan judul jelas, lahir dari percakapan yang sudah ada, bisa memiliki Dokumen terkait.

### 4.2 Asal-Usul Topik (Parent)

Topik selalu memiliki **parent** — konteks tempat topik itu dibuat:

| Parent | Anggota Topik | Contoh |
|---|---|---|
| **Chat Personal** | Kedua peserta chat (otomatis) | Andi & Budi membuat topik "Pembagian Lahan" dari chat personal mereka |
| **Chat Grup** | Sebagian atau seluruh member grup | Dari grup "Tim Proyek", Andi & Sari membuat topik "Desain UI" |

```
Chat Personal (Andi ↔ Budi)
├── Percakapan biasa
├── 📄 Dokumen "Catatan Hutang"
└── 📌 Topik "Pembagian Lahan"
    ├── Diskusi terfokus
    └── 📄 Dokumen "Kontrak Pembagian"

Grup "Tim Proyek" (Andi, Budi, Sari)
├── Percakapan grup
├── 📄 Dokumen "Brief Proyek"
├── 📌 Topik "Desain UI" (Andi & Sari)
│   ├── Diskusi terfokus
│   └── 📄 Dokumen "Wireframe v2"
└── 📌 Topik "Backend API" (Budi & Sari)
    ├── Diskusi terfokus
    └── 📄 Dokumen "API Spec"
```

### 4.3 Membuat Topik Baru

Topik dibuat dari dalam Chat Personal atau Grup:

| Field | Tipe | Wajib | Keterangan |
|---|---|---|---|
| Ikon Topik | Pilihan emoji | Ya | Identifikasi visual |
| Nama Topik | Teks bebas | Ya | Judul deskriptif, cth: "Rencana Panen Oktober", "Proyek Website" |
| Anggota | Multi-pilih dari anggota parent | Ya (min. 1) | Dari personal: otomatis keduanya. Dari grup: pilih sebagian/semua |
| Deskripsi | Teks bebas | Tidak | Penjelasan tujuan topik |

#### 4.3.1 Pilihan Ikon Topik

| Ikon | Konteks |
|---|---|
| 💬 | Umum / diskusi bebas |
| 🏡 | Rumah tangga |
| 🌾 | Pertanian dan kebun |
| 🏥 | Kesehatan |
| 📚 | Pendidikan |
| 💰 | Keuangan |
| 🛒 | Belanja |
| 📋 | Perencanaan / checklist |
| 💼 | Bisnis / pekerjaan |
| 🤝 | Kesepakatan / kontrak |

### 4.4 Fitur dalam Topik

Setiap Topik memiliki dua tab:

- **Tab Diskusi (💬):** Ruang chat antar anggota topik (sama seperti fitur chat) + dokumen inline.
- **Tab Dokumen (📄):** Daftar semua dokumen yang dimiliki topik ini.

Tombol 📄 di header topik berfungsi sebagai pintasan membuat dokumen baru yang langsung terhubung ke topik.

### 4.5 Aturan Keanggotaan Topik

- Pembuat topik otomatis menjadi admin dan tidak bisa dikeluarkan.
- Anggota topik **harus berasal dari parent** (anggota chat personal atau member grup).
- Topik dari chat personal otomatis berisi kedua peserta.
- Topik dari grup bisa berisi sebagian atau seluruh member grup.
- Admin bisa menambah anggota (dari member parent) atau mengeluarkan anggota.

---

## 5. Fitur Dokumen Kolaboratif (Notion-Style)

### 5.1 Konsep Dokumen

Dokumen di Chatat adalah **dokumen kolaboratif bergaya Notion** — jauh lebih kaya dari catatan teks biasa. Dokumen menggunakan sistem **block-based editor** di mana setiap elemen konten (paragraf, heading, tabel, checklist, dll.) adalah sebuah "block" yang bisa ditambah, dihapus, dipindahkan, dan diformat secara independen.

### 5.2 Dokumen Hidup di Mana Saja (Hybrid Model)

Dokumen menggunakan model **hybrid** — muncul di dua tempat sekaligus:

1. **Sebagai card inline di chat** — saat dokumen dibuat atau di-share, muncul sebagai card preview di dalam alur chat, memberikan konteks temporal (kapan dibuat/dibagikan).
2. **Di tab Dokumen** — dokumen juga otomatis muncul di tab Dokumen milik konteks tersebut (chat personal, grup, atau topik), memudahkan pencarian.

```
[Chat Personal: Andi ↔ Budi]
├── Tab Chat:
│   ├── 💬 Andi: "Kita perlu catat pembagian"
│   ├── 📄 [Card: Pembagian Hasil Panen]  ← inline preview, klik untuk buka
│   └── 💬 Budi: "Sudah saya tanda tangani"
│
└── Tab Dokumen:
    └── 📄 Pembagian Hasil Panen           ← juga muncul di sini
```

#### Ownership Dokumen per Konteks

| Konteks Pemilik | Siapa Bisa Akses | Contoh |
|---|---|---|
| **Chat Personal** | Kedua peserta chat | "Catatan Hutang" antara Andi & Budi |
| **Chat Grup** | Semua member grup | "Brief Proyek" di grup Tim Proyek |
| **Topik** (dari personal) | Anggota topik (kedua orang) | "Kontrak Pembagian" di topik terfokus |
| **Topik** (dari grup) | Anggota topik (subset member grup) | "Wireframe UI" di topik Desain |
| **Standalone** | Pemilik + kolaborator pilihan manual | Dokumen pribadi yang di-share manual |

> 💡 **Mengapa Hybrid?**
> Konteks temporal tetap ada (di chat), pengelolaan mudah (di tab). Dokumen tidak hilang meski tenggelam di chat, karena selalu bisa diakses dari tab Dokumen.

### 5.3 Tipe Block yang Didukung

#### 5.3.1 Block Teks

| Block | Markdown | Deskripsi |
|---|---|---|
| **Paragraf** | Teks biasa | Teks narasi standar |
| **Heading 1** | `# Judul` | Judul utama, ukuran besar |
| **Heading 2** | `## Sub-judul` | Sub-judul |
| **Heading 3** | `### Sub-sub-judul` | Sub-sub-judul |
| **Bold** | `**teks**` | Teks tebal |
| **Italic** | `*teks*` | Teks miring |
| **Strikethrough** | `~~teks~~` | Teks dicoret |
| **Inline Code** | `` `kode` `` | Teks kode inline |
| **Blockquote** | `> kutipan` | Kutipan/catatan penting dengan garis kiri |
| **Divider** | `---` | Garis pembatas horizontal |

#### 5.3.2 Block Daftar

| Block | Markdown | Deskripsi |
|---|---|---|
| **Bullet List** | `- item` | Daftar dengan poin bulat |
| **Numbered List** | `1. item` | Daftar bernomor otomatis |
| **Checklist** | `- [ ] item` | Daftar dengan kotak centang interaktif |

#### 5.3.3 Block Data & Media

| Block | Cara Akses | Deskripsi |
|---|---|---|
| **Tabel** | Ketik `/tabel` | Tabel dengan kolom dan baris dinamis (tambah/hapus kolom-baris) |
| **Callout** | Ketik `/callout` | Kotak info berwarna dengan ikon emoji — untuk catatan penting atau peringatan |
| **Code Block** | Ketik `/kode` | Area kode dengan syntax highlighting |
| **Toggle** | Ketik `/toggle` | Konten yang bisa dibuka/tutup (accordion) |

### 5.4 Slash Commands (`/`)

Pengguna mengetik `/` di baris kosong untuk memunculkan menu pilihan block. Ini adalah cara utama menambahkan block selain teks biasa.

| Command | Hasil |
|---|---|
| `/h1` atau `/heading1` | Heading 1 |
| `/h2` atau `/heading2` | Heading 2 |
| `/h3` atau `/heading3` | Heading 3 |
| `/bullet` atau `/poin` | Bullet list |
| `/angka` atau `/numbered` | Numbered list |
| `/centang` atau `/checklist` | Checklist |
| `/tabel` | Tabel baru |
| `/callout` | Callout box |
| `/kode` | Code block |
| `/toggle` | Toggle/accordion |
| `/pembatas` atau `/divider` | Garis pembatas |
| `/kutipan` atau `/quote` | Blockquote |

### 5.5 Fitur Tabel Lanjutan

Tabel di Chatat lebih kaya dari tabel markdown biasa:

| Fitur | Deskripsi |
|---|---|
| **Tambah/hapus kolom** | Klik tombol `+` di kanan header untuk tambah kolom |
| **Tambah/hapus baris** | Klik tombol `+ Baris` di bawah tabel |
| **Resize kolom** | Drag pembatas kolom untuk mengubah lebar |
| **Header row** | Baris pertama otomatis menjadi header dengan style berbeda |
| **Cell editing** | Klik langsung pada sel untuk mengedit konten |
| **Tipe kolom** | Teks, Angka, Tanggal, Checkbox (pilih saat buat kolom) |

### 5.6 Toolbar Formatting

Saat pengguna menyeleksi teks, muncul floating toolbar dengan opsi:

- **B** (Bold) — tekan `Ctrl/Cmd + B`
- *I* (Italic) — tekan `Ctrl/Cmd + I`
- ~~S~~ (Strikethrough) — tekan `Ctrl/Cmd + Shift + S`
- `<>` (Inline code)
- 🔗 (Tambah link)
- Highlight warna

### 5.7 Kolaborasi Dokumen

Setiap dokumen memiliki **pemilik** (pembuat) dan **kolaborator** (anggota yang dipilih):

| Peran | Hak Akses |
|---|---|
| **Pemilik** | Buat, edit, hapus, kunci, atur kolaborator |
| **Editor** | Edit konten dokumen, tambah block, isi tabel |
| **Viewer** | Hanya bisa melihat, tidak bisa mengedit |

### 5.8 Metadata Dokumen

| Field | Tipe | Keterangan |
|---|---|---|
| Judul | Teks | Judul utama dokumen |
| Ikon | Emoji | Ikon identifikasi dokumen |
| Cover | Pilihan warna/gradien | Banner dekoratif di atas dokumen (opsional) |
| Label/Tag | Multi-tag | Label bebas untuk kategorisasi (cth: "keuangan", "proyek-A", "mobil-avanza") |
| Kolaborator | Multi-pilih dari kontak | Siapa saja yang bisa mengakses |
| Konteks Parent | Auto-set | Chat personal, grup, atau topik tempat dokumen dibuat |
| Entitas/Tag Subjek | Multi-tag dinamis | Tag subjek spesifik (lihat 5.9) |

### 5.9 Entitas Dinamis (Entity Tags) 🏷️

Entitas adalah label dinamis yang bisa dibuat secara bebas oleh pengguna untuk menandai **subjek spesifik** dalam dokumen. Berbeda dengan kategori/label yang bersifat umum, Entitas mengidentifikasi **objek konkret** yang jadi pokok bahasan.

> 💡 **Entitas bisa berupa apa saja**
> Tidak ada batasan jenis entitas. Pengguna bebas membuat entitas sesuai kebutuhan — bisa lahan pertanian, kendaraan, anak, properti, proyek, hewan ternak, **atau bahkan orang dari daftar kontak**. Entitas berupa kontak memungkinkan dokumen langsung terkoneksi dengan orang yang bersangkutan.

#### Contoh Penggunaan Entitas

| Jenis Objek | Contoh Entitas | Dokumen Terkait |
|---|---|---|
| **Orang (Kontak)** | `Budi (kontak)`, `Sari (kontak)` | Hutang piutang, kesepakatan, riwayat kesehatan anak |
| Lahan | `Sawah Barat`, `Ladang Utara`, `Kebun Singkong` | Catatan panen, biaya pupuk, pembagian hasil |
| Kendaraan | `Avanza 2020`, `Motor Vario`, `Truk Angkut` | Service record, pajak, pemakaian BBM |
| Properti | `Rumah Jl. Melati`, `Kos-kosan B`, `Ruko Pasar` | Kontrak sewa, renovasi, pendapatan |
| Proyek | `Proyek Website`, `Renovasi Dapur`, `Acara Nikahan` | Timeline, anggaran, notulen |
| Hewan | `Sapi #1`, `Kambing Pejantan`, `Kolam Ikan A` | Catatan pakan, vaksin, penjualan |
| Perangkat | `Laptop Asus`, `HP Samsung`, `Mesin Cuci` | Garansi, service, spesifikasi |

> 🔗 **Entitas Kontak**
> Saat pengguna memilih kontak sebagai entitas, dokumen menjadi "bertag" dengan orang tersebut. Ini memudahkan pencarian — misal mencari semua dokumen yang terkait dengan `Budi`. Namun, tag kontak **tidak otomatis memberi akses** ke orang tersebut. Hak akses dokumen tetap mengikuti konteks tempat dokumen berada (chat personal, grup, atau topik). Jika dokumen ada di chat 1-on-1, hanya kedua peserta yang bisa melihatnya — meskipun entitasnya merujuk ke orang lain.

#### Cara Kerja Entitas

1. Saat membuat atau mengedit dokumen, pengguna menambahkan entitas di field "Entitas/Subjek".
2. Pengguna bisa **mengetik entitas baru** (otomatis tersimpan untuk penggunaan berikutnya), **memilih dari entitas yang pernah dibuat**, atau **memilih dari daftar kontak**.
3. Satu dokumen bisa memiliki **beberapa entitas** (misalnya: `Sawah Barat` + `Budi (kontak)`).
4. Entitas bisa digunakan sebagai **filter** di halaman Dokumen — cari semua dokumen yang terkait `Avanza 2020` atau `Budi`.
5. Entitas bersifat **global** — bisa digunakan di dokumen mana pun, tidak terikat pada satu konteks.
6. Entitas kontak memiliki **link langsung** ke profil orang tersebut di Chatat.

### 5.10 Riwayat Dokumen

Setiap dokumen memiliki log riwayat otomatis:

| Aksi | Keterangan Riwayat |
|---|---|
| Pembuatan | `"Dibuat oleh [Nama]"` |
| Pengeditan | `"Diedit oleh [Nama]"` — dengan timestamp |
| Kolaborator ditambah | `"[Nama] ditambahkan sebagai [peran]"` |
| Tanda tangan diberikan | `"[Nama] menandatangani dokumen"` |
| Dokumen dikunci | `"Dokumen dikunci — semua tanda tangan terkumpul"` |

### 5.11 Template Dokumen

Pengguna dapat memulai dengan template kosong atau memilih template bawaan:

| Template | Konten Default |
|---|---|
| **Kosong** | Dokumen kosong dengan judul saja |
| **Notulen Rapat** | Heading: Agenda, Peserta, Pembahasan, Keputusan |
| **Daftar Belanja** | Tabel: Nama Barang, Jumlah, Harga Satuan, Total |
| **Catatan Keuangan** | Tabel: Tanggal, Keterangan, Pemasukan, Pengeluaran, Saldo |
| **Catatan Kesehatan** | Heading: Keluhan, Diagnosis, Obat, Dokter, Kunjungan Berikutnya |
| **Kesepakatan Bersama** | Heading: Pihak, Isi Kesepakatan, Ketentuan, Area Tanda Tangan |
| **Catatan Pertanian** | Tabel: Lahan, Tanaman, Tanggal Tanam, Hasil Panen, Catatan |
| **Inventaris Aset** | Tabel: Nama Aset, Jenis, Lokasi, Kondisi, Catatan |

---

## 6. Penguncian Dokumen

### 6.1 Konsep Penguncian

Penguncian dokumen adalah fitur yang memungkinkan sebuah dokumen dijadikan **"dokumen final"** yang tidak bisa diubah lagi setelah dikunci. Terdapat dua mekanisme penguncian:

### 6.2 Penguncian Manual (oleh Pemilik)

Pemilik dokumen dapat mengunci dokumen kapan saja tanpa memerlukan tanda tangan. Ini berguna untuk dokumen yang sudah final dan tidak perlu persetujuan pihak lain.

| Langkah | Aksi |
|---|---|
| 1 | Pemilik membuka dokumen |
| 2 | Menekan menu `⋮` → "Kunci Dokumen" |
| 3 | Konfirmasi penguncian |
| 4 | Dokumen terkunci permanen — tidak bisa diedit |

### 6.3 Penguncian dengan Tanda Tangan Digital (Kontrak)

Untuk dokumen yang membutuhkan persetujuan bersama:

| Langkah | Aktor | Aksi | Status |
|---|---|---|---|
| 1 | Pemilik dokumen | Mengaktifkan "Butuh tanda tangan" | Draft |
| 2 | Pemilik dokumen | Memilih siapa saja yang harus menandatangani dari kontak | Menunggu Tanda Tangan |
| 3 | Pemilik dokumen | Menyimpan dokumen | Aktif — badge ✍️ muncul |
| 4 | Penandatangan A | Membuka dokumen, review isi | Menunggu |
| 5 | Penandatangan A | Menekan "Tandatangani Sekarang" | 1 dari N sudah |
| 6 | Penandatangan B, C... | Mengulangi langkah 4-5 | Bertambah |
| 7 | Sistem | Otomatis mengunci saat semua sudah tandatangan | 🔒 TERKUNCI PERMANEN |

### 6.4 Tampilan Status Tanda Tangan

| Status | Tampilan | Keterangan |
|---|---|---|
| Belum menandatangani | ⏳ Menunggu (abu-abu) | Belum memberikan tanda tangan |
| Sudah menandatangani | ✅ Ditandatangani · [Tanggal] (hijau) | Timestamp kapan ditandatangani |
| Semua selesai | Banner hijau "Dokumen Terkunci" | Otomatis terkunci permanen |

### 6.5 Badge Visual pada Kartu Dokumen

| Badge | Kondisi | Arti |
|---|---|---|
| ✍️ Menunggu Tanda Tangan (ungu) | Ada tanda tangan yang belum terkumpul | Perlu aksi dari penandatangan |
| 🔒 Terkunci (kuning) | Dokumen sudah dikunci (manual atau via tanda tangan) | Final, tidak bisa diubah |
| 📄 Draft (abu) | Dokumen belum dikunci | Masih bisa diedit |

### 6.6 Aturan Penguncian

- Setelah terkunci, **tidak ada yang bisa mengedit** isi dokumen — termasuk pemilik.
- Dokumen terkunci tetap bisa **dilihat** oleh semua kolaborator.
- Penguncian bersifat **permanen** — tidak bisa dibuka kembali.
- Riwayat penguncian tercatat di log dokumen.
- Dokumen yang belum dikunci bisa dihapus oleh pemilik, yang sudah dikunci tidak bisa dihapus.

### 6.7 Contoh Penggunaan

| Skenario | Isi Dokumen | Penandatangan |
|---|---|---|
| Pembagian hasil panen | Tabel pembagian per lahan, persentase per anggota | Andi, Budi |
| Kesepakatan sewa kendaraan | Detail sewa Mobil Avanza, durasi, biaya | Andi, Sari |
| Kontrak proyek bersama | Scope pekerjaan, timeline, pembayaran | Semua anggota tim |
| Notulen rapat organisasi | Agenda, pembahasan, keputusan — dikunci setelah selesai | Semua peserta |
| Daftar inventaris | Daftar aset dan kondisi — dikunci manual oleh pemilik | - (kunci manual) |

---

## 7. Navigasi & Antarmuka

### 7.1 Navigasi Utama (Bottom Navigation)

Navigasi utama mengikuti pola WhatsApp dengan dua tab di bagian bawah layar:

| Tab | Ikon | Konten |
|---|---|---|
| **Chat** | 💬 | Daftar semua chat personal dan grup. Topik diakses dari dalam chat/grup. |
| **Dokumen** | 📄 | Daftar semua dokumen kolaboratif (lintas semua konteks) |

### 7.2 Header & Aksi Cepat

Header aplikasi menampilkan:
- **Kiri:** Logo/nama "Chatat"
- **Kanan:** Ikon pencarian 🔍, ikon profil (avatar pengguna)

### 7.3 Tombol Aksi Utama (FAB)

Tombol bulat **(+)** berwarna hijau di pojok kanan bawah. Fungsinya berubah sesuai tab aktif:

| Tab Aktif | Aksi FAB |
|---|---|
| Chat | Buka daftar kontak untuk memulai chat baru / buat grup baru |
| Dokumen | Buat dokumen baru standalone (pilih template atau kosong) |

### 7.4 Halaman Kontak

Dapat diakses dari FAB di tab Chat atau dari menu. Menampilkan:
- Daftar semua pengguna terdaftar yang ada di kontak ponsel
- Avatar, nama, dan status masing-masing
- Tap pada kontak untuk memulai/membuka chat personal
- Tombol "Buat Grup" di atas daftar
- Field pencarian untuk mencari berdasarkan nama atau nomor HP

### 7.5 Filter Dokumen

Pada tab Dokumen, terdapat baris filter horizontal:

- 🗂️ **Semua** — seluruh dokumen
- 🔒 **Dikunci** — dokumen yang sudah terkunci
- ✍️ **Menunggu Tanda Tangan** — dokumen yang perlu ditandatangani
- 📄 **Draft** — dokumen yang masih bisa diedit
- 🏷️ **Per Label** — filter berdasarkan tag/label
- 📌 **Per Entitas** — filter berdasarkan subjek/entitas (cth: semua dokumen tentang "Avanza 2020")

### 7.6 Pencarian Global

Ikon pencarian di header memungkinkan pencarian lintas fitur:
- Cari pesan dalam chat
- Cari kontak berdasarkan nama/nomor
- Cari dokumen berdasarkan judul atau konten
- Cari topik berdasarkan nama

---

## 8. Penyimpanan Data

### 8.1 Arsitektur Penyimpanan (WhatsApp-Style)

Penyimpanan data mengikuti model WhatsApp — **local-first** dengan server sebagai relay dan sync:

| Komponen | Lokasi | Penjelasan |
|---|---|---|
| **Pesan Chat & Topik** | Lokal (device) | Disimpan di database lokal perangkat (SQLite/Realm). Server hanya relay — pesan dikirim ke server, diteruskan ke penerima, lalu dihapus dari server setelah terkirim. |
| **Dokumen Kolaboratif** | Server | Karena bersifat kolaboratif (banyak editor), dokumen disimpan di server dan di-sync ke semua kolaborator. |
| **Profil & Kontak** | Server + Lokal | Profil pengguna di server. Daftar kontak di-cache lokal. |
| **Entitas** | Server + Lokal | Entitas milik pengguna disimpan di server, di-cache lokal. |
| **Media (foto, file)** | Server + Lokal | Media di-upload ke server, di-download oleh penerima. Tersimpan lokal setelah diunduh. |
| **Backup** | Cloud (opsional) | Pengguna bisa backup chat ke Google Drive (Android) atau iCloud (iOS), seperti WhatsApp. |

#### Prinsip Utama

- **Store-and-forward**: Server menyimpan pesan sementara sampai terkirim, lalu menghapusnya
- **Lokal sebagai sumber utama**: Riwayat chat ada di perangkat pengguna, bukan di cloud
- **Enkripsi end-to-end**: Pesan hanya bisa dibaca oleh pengirim dan penerima
- **Status pengiriman**: ✓ terkirim ke server, ✓✓ terkirim ke penerima, biru = dibaca

### 8.2 Struktur Data Utama

#### 8.2.1 Struktur Data Pengguna (`User`)

| Field | Tipe | Keterangan |
|---|---|---|
| `id` | String | ID unik pengguna |
| `name` | String | Nama tampil |
| `phone` | String | Nomor HP untuk verifikasi (format internasional) |
| `avatar` | String (emoji) | Identitas visual |
| `status` | String | Status/bio singkat |
| `lastSeen` | ISO DateTime | Waktu terakhir aktif |
| `createdAt` | ISO DateTime | Waktu pendaftaran |

#### 8.2.2 Struktur Data Chat (`Chat`)

| Field | Tipe | Keterangan |
|---|---|---|
| `id` | String | ID unik chat |
| `type` | String (enum) | `personal` / `group` |
| `name` | String \| null | Nama grup (null untuk personal) |
| `icon` | String \| null | Ikon emoji grup (null untuk personal) |
| `memberIds` | Array String | ID semua anggota chat |
| `adminIds` | Array String | ID admin grup (pembuat) |
| `messages` | Array Object | Pesan: `{id, senderId, text, replyTo, at, readBy}` |
| `pinnedAt` | ISO DateTime \| null | Waktu di-pin (null jika tidak di-pin) |
| `createdAt` | ISO DateTime | Waktu pembuatan |

#### 8.2.3 Struktur Data Topik (`Topic`)

| Field | Tipe | Keterangan |
|---|---|---|
| `id` | String | ID unik topik |
| `name` | String | Nama topik |
| `icon` | String (emoji) | Ikon visual topik |
| `description` | String | Deskripsi tujuan topik |
| `parentType` | String (enum) | `personal` / `group` — dari mana topik dibuat |
| `parentId` | String | ID chat personal atau grup parent |
| `memberIds` | Array String | ID semua anggota topik |
| `adminIds` | Array String | ID admin topik |
| `messages` | Array Object | Pesan diskusi: `{id, senderId, text, replyTo, at}` |
| `createdAt` | ISO DateTime | Waktu pembuatan |

#### 8.2.4 Struktur Data Dokumen (`Document`)

| Field | Tipe | Keterangan |
|---|---|---|
| `id` | String | ID unik dokumen |
| `title` | String | Judul dokumen |
| `icon` | String (emoji) | Ikon dokumen |
| `cover` | String \| null | Kode warna/gradien cover |
| `blocks` | Array Object | Konten block-based (lihat 8.2.5) |
| `tags` | Array String | Label/tag untuk kategorisasi |
| `entities` | Array String | Entitas/subjek terkait (lahan, kendaraan, anak, dsb) |
| `ownerId` | String | ID pemilik dokumen |
| `collaborators` | Array Object | `{userId, role}` — role: `editor` / `viewer` |
| `topicId` | String \| null | ID topik terkait (jika dibuat dari topik) |
| `chatId` | String \| null | ID chat personal terkait (jika dibuat dari chat personal) |
| `groupId` | String \| null | ID grup terkait (jika dibuat dari grup) |
| `requireSigs` | Boolean | Apakah membutuhkan tanda tangan |
| `signerIds` | Array String | ID yang harus menandatangani |
| `sigs` | Object | Map: `{userId: {at, name}}` — tanda tangan yang masuk |
| `locked` | Boolean | Apakah dokumen sudah terkunci |
| `lockedAt` | ISO DateTime \| null | Waktu penguncian |
| `lockedBy` | String \| null | `manual` / `signatures` |
| `history` | Array Object | Log riwayat: `[{at, action, userId}]` |
| `createdAt` | ISO DateTime | Waktu pembuatan |
| `updatedAt` | ISO DateTime | Waktu terakhir diperbarui |

#### 8.2.5 Struktur Block Dokumen

Setiap block dalam array `blocks` memiliki struktur:

| Field | Tipe | Keterangan |
|---|---|---|
| `id` | String | ID unik block |
| `type` | String (enum) | `paragraph`, `heading1`, `heading2`, `heading3`, `bullet-list`, `numbered-list`, `checklist`, `table`, `callout`, `code`, `toggle`, `divider`, `quote` |
| `content` | String \| null | Konten teks (dengan formatting markdown inline) |
| `checked` | Boolean \| null | Status centang (khusus checklist) |
| `children` | Array Block \| null | Sub-block (khusus toggle) |
| `rows` | Array Array \| null | Data tabel (khusus table) |
| `columns` | Array Object \| null | Definisi kolom: `{name, type}` (khusus table) |
| `language` | String \| null | Bahasa kode (khusus code block) |
| `emoji` | String \| null | Ikon callout (khusus callout) |
| `color` | String \| null | Warna callout/highlight |

---

## 9. Desain Visual & Antarmuka

### 9.1 Prinsip Desain

| Prinsip | Implementasi |
|---|---|
| Mobile-First | Layout maksimal 430px lebar, elemen dioptimalkan untuk sentuhan jari |
| Dark Theme | Latar belakang gelap (#0F1117) untuk kenyamanan mata |
| WhatsApp-Familiar | Layout dan pattern interaksi mengikuti kebiasaan pengguna WhatsApp |
| Kontras Tinggi | Teks putih di atas latar gelap, aksen hijau untuk elemen utama |
| Minimal Kognitif | Satu tugas per layar, navigasi sesederhana mungkin |
| Umpan Balik Visual | Animasi transisi, typing indicator, status pesan untuk konfirmasi aksi |

### 9.2 Palet Warna

| Nama | Kode Hex | Penggunaan |
|---|---|---|
| Background | `#0F1117` | Latar belakang utama |
| Surface | `#1A1D27` | Kartu, header, navigasi bawah |
| Surface 2 | `#222637` | Elemen nested, bubble chat orang lain |
| Border | `#2E3348` | Pembatas antar elemen |
| Teks Utama | `#E8EAF0` | Semua teks utama |
| Teks Muted | `#6B7280` | Label sekunder, timestamp, metadata |
| Aksen Hijau | `#6EE7B7` | CTA utama, bubble chat sendiri, status online |
| Aksen Ungu | `#818CF8` | Nama pengirim di grup, badge tanda tangan |
| Bahaya | `#F87171` | Error, logout, hapus |
| Peringatan | `#FBBF24` | Badge dokumen terkunci |
| Aksen Biru | `#60A5FA` | Link, judul dokumen |

### 9.3 Tipografi

| Jenis | Font | Penggunaan |
|---|---|---|
| Font Utama UI | Plus Jakarta Sans | Semua teks antarmuka: judul, label, tombol, navigasi |
| Font Dokumen | Inter | Konten dokumen untuk keterbacaan optimal |
| Font Kode | JetBrains Mono | Code block dalam dokumen |

---

## 10. Roadmap Pengembangan Masa Depan

### 10.1 Fitur Prioritas Tinggi

| Fitur | Deskripsi | Alasan Prioritas |
|---|---|---|
| Foto & Media | Kirim foto, video di chat dan embed di dokumen | Komunikasi lebih kaya |
| Notifikasi Push | Notifikasi pesan baru dan tanda tangan pending | Responsivitas real-time |
| Voice Message | Kirim pesan suara di chat (WhatsApp-style) | Kemudahan untuk input panjang |
| Ekspor PDF | Ekspor dokumen terkunci menjadi PDF | Arsip digital/fisik |
| Real-time Sync | Sinkronisasi chat dan dokumen real-time antar perangkat | Pengalaman kolaborasi lebih baik |

### 10.2 Fitur Prioritas Menengah

- **Mention** (`@nama`) dalam chat dan dokumen.
- **Reaction emoji** pada pesan chat.
- **Version history** dokumen — lihat dan kembalikan versi sebelumnya.
- **Comment** pada block dokumen — diskusi inline tanpa mengubah konten.
- **Kalender** — tampilkan dokumen dan event berdasarkan tanggal.
- **Pengingat/alarm** untuk deadline dan jadwal.
- **Status pesan** — online, typing, terakhir dilihat.

### 10.3 Fitur Jangka Panjang

- **Panggilan suara/video** antar pengguna.
- **Workspace/Organisasi** — satu akun bergabung ke beberapa workspace (keluarga, tim kerja, komunitas).
- **Database Notion-style** — tabel sebagai database dengan views (table, board, calendar).
- **Backup & restore** ke Google Drive / iCloud.
- **Enkripsi end-to-end** untuk semua data.
- **Bot/reminder otomatis** berdasarkan jadwal dan event.
- **Widget** di home screen smartphone untuk akses cepat.

---

## 11. Glosarium

| Istilah | Definisi |
|---|---|
| **Chat Personal** | Percakapan 1-on-1 antara dua pengguna, seperti chat WhatsApp biasa |
| **Chat Grup** | Percakapan dengan beberapa anggota, bisa dibuat oleh siapa saja |
| **Kontak** | Pengguna Chatat yang nomor HP-nya ada di kontak ponsel — terkoneksi otomatis |
| **Topik** | Ruang diskusi terfokus dengan tujuan spesifik, terpisah dari chat biasa |
| **Dokumen** | Dokumen kolaboratif Notion-style dengan block-based editor |
| **Block** | Unit konten dalam dokumen: paragraf, heading, tabel, checklist, dll |
| **Slash Command** | Mengetik `/` untuk memunculkan menu pilihan block dalam editor dokumen |
| **Penguncian** | Aksi mengunci dokumen secara permanen agar tidak bisa diedit lagi |
| **Tanda Tangan Digital** | Persetujuan resmi oleh pengguna, tercatat dengan timestamp |
| **Kolaborator** | Anggota yang memiliki akses ke dokumen dengan peran tertentu (editor/viewer) |
| **Template** | Struktur dokumen siap pakai yang bisa dipilih saat membuat dokumen baru |
| **FAB** | Floating Action Button — tombol aksi utama bulat di pojok kanan bawah |
| **Badge** | Label kecil visual yang menunjukkan status (unread, pending, terkunci) |
| **Sesi** | Status login aktif yang tersimpan agar tidak perlu login ulang |
| **Avatar** | Emoji identitas visual setiap pengguna |
| **Bottom Sheet** | Modal formulir yang muncul dari bawah layar (standar UI mobile) |
| **Entitas** | Subjek dinamis yang bisa ditag dalam dokumen — bisa berupa apa saja: lahan, kendaraan, anak, proyek, aset, dll |

---

*— Dokumen ini disiapkan sebagai referensi pengembangan. Versi 3.0 | Chatat © 2026 —*
