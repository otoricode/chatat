# Git Workflow

> Alur kerja Git untuk tim development Chatat.
> Menggunakan trunk-based development dengan short-lived feature branches.

---

## Branch Strategy

```
main ─────────────────────────────────────────────────>
  │                     │                │
  ├── feature/document-locking ── PR ── merge
  │                     │
  ├── fix/otp-timeout ─────────── PR ── merge
  │                                     │
  └── release/1.0.0 ─── tag v1.0.0 ─── merge
```

### Branch Types

| Branch | Pattern | Lifetime | From | Merge To |
|--------|---------|----------|------|----------|
| Main | `main` | Permanent | — | — |
| Feature | `feature/{desc}` | 1-5 hari | `main` | `main` |
| Fix | `fix/{desc}` | 1-2 hari | `main` | `main` |
| Refactor | `refactor/{desc}` | 1-3 hari | `main` | `main` |
| Release | `release/{version}` | 1-2 hari | `main` | `main` + tag |
| Hotfix | `hotfix/{desc}` | < 1 hari | latest tag | `main` + tag |

### Rules:
- Branch dari `main`, merge ke `main`
- Short-lived branches (max 5 hari)
- Rebase sebelum merge (linear history)
- Delete branch setelah merge

---

## Commit Messages

### Format: Conventional Commits

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

| Type | Description | Example |
|------|-------------|---------|
| `feat` | Fitur baru | `feat(chat): add typing indicator` |
| `fix` | Bug fix | `fix(auth): handle expired OTP` |
| `refactor` | Refactoring tanpa change behavior | `refactor(ws): extract hub from handler` |
| `docs` | Dokumentasi | `docs: update API reference` |
| `test` | Menambah atau fix test | `test(document): add lock edge cases` |
| `perf` | Performance improvement | `perf(chat): batch message inserts` |
| `style` | Formatting | `style: apply gofmt` |
| `chore` | Build, CI, dependencies | `chore: update go to 1.23` |
| `ci` | CI/CD changes | `ci: add go test workflow` |

### Scopes

| Scope | Area |
|-------|------|
| `auth` | Authentication (OTP, token) |
| `chat` | Chat personal & group |
| `topic` | Topic system |
| `document` | Document & blocks |
| `entity` | Entity/tag system |
| `contact` | Contact management |
| `ws` | WebSocket |
| `media` | Media upload/download |
| `notif` | Push notifications |
| `ui` | General UI (mobile) |
| `db` | Database |
| `api` | API endpoints |
| `deps` | Dependencies |

### Examples

```
feat(chat): implement group chat creation

Add group creation with member selection and icon picker.
Minimum 3 members required.

Closes #42

---

fix(document): prevent concurrent lock requests

Use database-level locking to prevent race condition when
two users try to lock the same document simultaneously.

---

refactor(ws): split hub into separate read/write goroutines

Improve WebSocket scalability by separating read and write
loops per client connection.
```

### Rules:
- Lowercase description (no capital, no period)
- Imperative mood ("add" not "added" or "adds")
- Max 72 characters for subject line
- Body wraps at 80 characters
- Reference issues in footer: `Closes #42`, `Fixes #17`

---

## Pull Request Process

### PR Template

```markdown
## What

Brief description of what this PR does.

## Why

Why is this change needed.

## How

Technical approach / key decisions.

## Testing

- [ ] Unit tests added/updated
- [ ] Integration tests (if applicable)
- [ ] Manual testing done on iOS
- [ ] Manual testing done on Android

## Checklist

- [ ] Code follows style guide
- [ ] No `panic()` in production code
- [ ] No `any` types in TypeScript
- [ ] Error handling is proper
- [ ] Documentation updated (if needed)
```

### PR Rules:
- PR title follows commit convention: `feat(scope): description`
- Small PRs (< 400 lines changed preferred)
- Self-review sebelum request review
- All CI checks harus pass
- Squash merge to main (clean history)

---

## Release Process

### Versioning: Semantic Versioning

```
MAJOR.MINOR.PATCH
  │      │     │
  │      │     └── Bug fixes, no new features
  │      └──────── New features, backward compatible
  └─────────────── Breaking changes
```

### Release Steps

```bash
# 1. Create release branch
git checkout -b release/1.2.0

# 2. Update version in:
#    - server: internal/config/version.go
#    - mobile: app.json / package.json

# 3. Update CHANGELOG.md

# 4. Commit
git commit -m "chore: bump version to 1.2.0"

# 5. PR to main

# 6. After merge, tag
git tag -a v1.2.0 -m "Release v1.2.0"
git push origin v1.2.0

# 7. CI builds and publishes release
```

---

## CHANGELOG Format

```markdown
# Changelog

## [1.2.0] - 2025-03-15

### Added
- Group chat creation with member picker (#42)
- Document locking with signature collection (#55)
- Arabic language support (#60)

### Changed
- Improved WebSocket reconnection strategy (#58)
- Updated OTP expiry from 5min to 10min (#61)

### Fixed
- Message ordering bug in slow networks (#50)
- Profile photo upload crash on Android (#52)
```

---

## .gitignore

```gitignore
# Go
server/tmp/
server/bin/

# Node / React Native
mobile/node_modules/
mobile/.expo/
mobile/ios/Pods/
mobile/android/.gradle/
mobile/android/app/build/

# Environment
.env
.env.local
.env.*.local

# IDE
.idea/
.vscode/settings.json
*.swp

# OS
.DS_Store
Thumbs.db

# Build
*.apk
*.aab
*.ipa

# Test
coverage/
*.lcov

# Database
*.db
*.db-journal
```

---

## Workflow Summary

```
1. Buat branch dari main
   git checkout -b feature/document-locking

2. Develop + commit (conventional commits)
   git commit -m "feat(document): add lock mechanism"

3. Push + buat PR
   git push -u origin feature/document-locking

4. CI checks (test + lint + build)

5. Code review

6. Squash merge ke main

7. Delete branch

8. Repeat
```
