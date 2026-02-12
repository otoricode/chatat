# Development Rules & Standards

> **Project:** Chatat
> **Stack:** Go (Backend) + React Native (Frontend) + TypeScript
> **Last Updated:** 2025

---

## Quick Navigation

| Dokumen | Isi |
|---------|-----|
| [project-structure.md](project-structure.md) | Struktur folder dan file conventions |
| [go-style-guide.md](go-style-guide.md) | Go coding style, patterns, dan conventions |
| [react-native-style-guide.md](react-native-style-guide.md) | React Native/TypeScript coding style dan conventions |
| [design-patterns.md](design-patterns.md) | Design patterns yang digunakan di project |
| [naming-conventions.md](naming-conventions.md) | Penamaan file, variabel, fungsi, types |
| [error-handling.md](error-handling.md) | Error handling strategy di Go dan TypeScript |
| [testing-strategy.md](testing-strategy.md) | Testing approach, tools, dan coverage targets |
| [git-workflow.md](git-workflow.md) | Git branching, commit messages, PR process |

---

## Core Principles

1. **Type Safety First** — Semua code strongly typed. No `any` di TypeScript. Proper types di Go.
2. **Separation of Concerns** — Backend (Go) handle logic + data. Frontend (React Native) handle UI only.
3. **Error as Values** — Gunakan error returns di Go, bukan panic. Typed errors di TypeScript.
4. **Immutability by Default** — Prefer immutable data. Mutations explicit dan minimal.
5. **Single Responsibility** — Setiap module/function punya satu tugas jelas.
6. **Test What Matters** — Focus testing pada business logic dan edge cases.
7. **Performance Conscious** — Profile sebelum optimize. Jangan premature optimization.
8. **Documentation as Code** — Kode yang butuh komentar berarti perlu direfactor. Dokumentasi di doc comments.
9. **Mobile-First** — Semua keputusan desain dan performa dioptimalkan untuk perangkat mobile.
10. **Offline-Ready** — Data lokal sebagai sumber utama, sinkronisasi saat online.
