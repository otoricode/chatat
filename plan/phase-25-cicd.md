# Phase 25: CI/CD Pipeline

> Setup continuous integration dan continuous deployment.
> GitHub Actions untuk automated testing, building, dan deployment.

**Estimasi:** 3 hari
**Dependency:** Phase 24 (Testing)
**Output:** Fully automated CI/CD pipeline.

---

## Task 25.1: Go Backend CI

**Input:** Phase 24 test suite
**Output:** GitHub Actions workflow untuk backend

### Steps:
1. Buat `.github/workflows/backend.yml`:
   ```yaml
   name: Backend CI

   on:
     push:
       branches: [main, develop]
       paths: ['server/**']
     pull_request:
       branches: [main]
       paths: ['server/**']

   jobs:
     lint:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4
         - uses: actions/setup-go@v5
           with:
             go-version: '1.23'
         - name: golangci-lint
           uses: golangci/golangci-lint-action@v4
           with:
             working-directory: server
             version: latest

     test:
       runs-on: ubuntu-latest
       needs: lint
       services:
         postgres:
           image: postgres:16
           env:
             POSTGRES_USER: test
             POSTGRES_PASSWORD: test
             POSTGRES_DB: chatat_test
           ports: ['5432:5432']
           options: >-
             --health-cmd pg_isready
             --health-interval 10s
             --health-timeout 5s
             --health-retries 5
         redis:
           image: redis:7
           ports: ['6379:6379']
           options: >-
             --health-cmd "redis-cli ping"
             --health-interval 10s
             --health-timeout 5s
             --health-retries 5
       steps:
         - uses: actions/checkout@v4
         - uses: actions/setup-go@v5
           with:
             go-version: '1.23'
         - name: Run migrations
           run: |
             cd server
             go run cmd/migrate/main.go up
           env:
             DATABASE_URL: postgres://test:test@localhost:5432/chatat_test?sslmode=disable
         - name: Run tests
           run: |
             cd server
             go test ./... -v -race -coverprofile=coverage.out
           env:
             DATABASE_URL: postgres://test:test@localhost:5432/chatat_test?sslmode=disable
             REDIS_URL: redis://localhost:6379/15
         - name: Upload coverage
           uses: codecov/codecov-action@v4
           with:
             file: server/coverage.out
             flags: backend

     build:
       runs-on: ubuntu-latest
       needs: test
       steps:
         - uses: actions/checkout@v4
         - uses: actions/setup-go@v5
           with:
             go-version: '1.23'
         - name: Build
           run: |
             cd server
             CGO_ENABLED=0 go build -o chatat-server cmd/server/main.go
         - name: Upload artifact
           uses: actions/upload-artifact@v4
           with:
             name: server-binary
             path: server/chatat-server
   ```
2. golangci-lint configuration:
   ```yaml
   # server/.golangci.yml
   linters:
     enable:
       - errcheck
       - gosimple
       - govet
       - ineffassign
       - staticcheck
       - unused
       - gosec
       - bodyclose
       - contextcheck
       - nilerr
       - sqlclosecheck
     disable:
       - deadcode
       - structcheck
       - varcheck

   linters-settings:
     gosec:
       excludes:
         - G304 # file path not tainted

   issues:
     exclude-rules:
       - path: _test\.go
         linters: [gosec, errcheck]
   ```

### Acceptance Criteria:
- [ ] Lint job: golangci-lint pass
- [ ] Test job: all tests pass with race detector
- [ ] Coverage uploaded to Codecov
- [ ] Build job: produces binary
- [ ] PostgreSQL + Redis services in CI
- [ ] Runs on push to main/develop

### Testing:
- [ ] Push to branch → workflow triggers
- [ ] Test failures block merge
- [ ] Coverage appears in PR

---

## Task 25.2: React Native CI

**Input:** Phase 24 test suite
**Output:** GitHub Actions workflow untuk mobile app

### Steps:
1. Buat `.github/workflows/mobile.yml`:
   ```yaml
   name: Mobile CI

   on:
     push:
       branches: [main, develop]
       paths: ['mobile/**']
     pull_request:
       branches: [main]
       paths: ['mobile/**']

   jobs:
     lint-and-typecheck:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4
         - uses: actions/setup-node@v4
           with:
             node-version: '20'
             cache: 'yarn'
             cache-dependency-path: mobile/yarn.lock
         - name: Install dependencies
           run: |
             cd mobile
             yarn install --frozen-lockfile
         - name: TypeScript check
           run: |
             cd mobile
             yarn tsc --noEmit
         - name: ESLint
           run: |
             cd mobile
             yarn lint

     test:
       runs-on: ubuntu-latest
       needs: lint-and-typecheck
       steps:
         - uses: actions/checkout@v4
         - uses: actions/setup-node@v4
           with:
             node-version: '20'
             cache: 'yarn'
             cache-dependency-path: mobile/yarn.lock
         - name: Install dependencies
           run: |
             cd mobile
             yarn install --frozen-lockfile
         - name: Run tests
           run: |
             cd mobile
             yarn test --coverage --watchAll=false
         - name: Upload coverage
           uses: codecov/codecov-action@v4
           with:
             file: mobile/coverage/lcov.info
             flags: mobile

     build-android:
       runs-on: ubuntu-latest
       needs: test
       if: github.ref == 'refs/heads/main' || github.ref == 'refs/heads/develop'
       steps:
         - uses: actions/checkout@v4
         - uses: actions/setup-java@v4
           with:
             distribution: 'temurin'
             java-version: '17'
         - uses: actions/setup-node@v4
           with:
             node-version: '20'
             cache: 'yarn'
             cache-dependency-path: mobile/yarn.lock
         - name: Install dependencies
           run: |
             cd mobile
             yarn install --frozen-lockfile
         - name: Build Android (Debug)
           run: |
             cd mobile/android
             ./gradlew assembleDebug
         - name: Upload APK
           uses: actions/upload-artifact@v4
           with:
             name: android-debug-apk
             path: mobile/android/app/build/outputs/apk/debug/app-debug.apk

     build-ios:
       runs-on: macos-latest
       needs: test
       if: github.ref == 'refs/heads/main' || github.ref == 'refs/heads/develop'
       steps:
         - uses: actions/checkout@v4
         - uses: actions/setup-node@v4
           with:
             node-version: '20'
             cache: 'yarn'
             cache-dependency-path: mobile/yarn.lock
         - name: Install dependencies
           run: |
             cd mobile
             yarn install --frozen-lockfile
         - name: Install pods
           run: |
             cd mobile/ios
             pod install
         - name: Build iOS (Debug)
           run: |
             cd mobile/ios
             xcodebuild -workspace Chatat.xcworkspace \
               -scheme Chatat \
               -configuration Debug \
               -sdk iphonesimulator \
               -destination 'platform=iOS Simulator,name=iPhone 15' \
               build
   ```

### Acceptance Criteria:
- [ ] TypeScript check passes
- [ ] ESLint passes
- [ ] All tests pass
- [ ] Android debug build succeeds
- [ ] iOS debug build succeeds
- [ ] Coverage uploaded
- [ ] Artifacts uploaded

### Testing:
- [ ] Push to branch → workflow triggers
- [ ] Build artifacts downloadable
- [ ] Coverage report in PR

---

## Task 25.3: Code Quality Gates

**Input:** Task 25.1, 25.2
**Output:** PR quality requirements

### Steps:
1. Branch protection rules:
   ```
   main branch:
   - Require PR reviews: 1
   - Require status checks:
     - backend / lint
     - backend / test
     - mobile / lint-and-typecheck
     - mobile / test
   - Require linear history
   - No force pushes
   ```
2. PR template:
   ```markdown
   <!-- .github/pull_request_template.md -->
   ## Summary
   <!-- Brief description of changes -->

   ## Type
   - [ ] Feature
   - [ ] Bug fix
   - [ ] Refactor
   - [ ] Documentation

   ## Checklist
   - [ ] Tests added/updated
   - [ ] Documentation updated
   - [ ] No console.log / fmt.Println left
   - [ ] i18n strings added for new UI text
   - [ ] Tested on iOS simulator
   - [ ] Tested on Android emulator
   ```
3. Commit message enforcement:
   ```yaml
   # .github/workflows/commitlint.yml
   name: Commit Lint
   on: [pull_request]
   jobs:
     commitlint:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4
           with:
             fetch-depth: 0
         - uses: wagoid/commitlint-github-action@v5
   ```
4. Commit convention:
   ```
   feat(scope): description    → new feature
   fix(scope): description     → bug fix
   perf(scope): description    → performance
   refactor(scope): description → refactor
   test(scope): description    → tests
   docs(scope): description    → documentation
   chore(scope): description   → maintenance
   
   Scopes: auth, chat, group, topic, doc, entity, editor, search,
           notif, i18n, backup, settings, security, perf, ci
   ```

### Acceptance Criteria:
- [ ] Branch protection on main
- [ ] Required status checks
- [ ] PR template available
- [ ] Commit lint enforced
- [ ] Conventional commits followed

### Testing:
- [ ] Create PR → checks run
- [ ] Bad commit message → rejected
- [ ] Failed test → PR blocked

---

## Task 25.4: Deployment Pipeline

**Input:** Task 25.1, 25.2
**Output:** Automated deployment for backend

### Steps:
1. Backend deployment (Docker):
   ```dockerfile
   # server/Dockerfile
   FROM golang:1.23-alpine AS builder
   WORKDIR /app
   COPY go.mod go.sum ./
   RUN go mod download
   COPY . .
   RUN CGO_ENABLED=0 go build -o chatat-server cmd/server/main.go

   FROM alpine:3.19
   RUN apk --no-cache add ca-certificates tzdata
   WORKDIR /app
   COPY --from=builder /app/chatat-server .
   COPY --from=builder /app/migrations ./migrations
   EXPOSE 8080
   CMD ["./chatat-server"]
   ```
2. Deploy workflow:
   ```yaml
   # .github/workflows/deploy.yml
   name: Deploy
   on:
     push:
       tags: ['v*']
   jobs:
     deploy-backend:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4
         - name: Build Docker image
           run: docker build -t chatat-server:${{ github.ref_name }} server/
         - name: Push to registry
           run: |
             docker tag chatat-server:${{ github.ref_name }} $REGISTRY/chatat-server:${{ github.ref_name }}
             docker push $REGISTRY/chatat-server:${{ github.ref_name }}
         # Deploy to production server (SSH or cloud provider)
   ```
3. Database migration in deployment:
   ```yaml
   - name: Run migrations
     run: |
       docker run --rm \
         -e DATABASE_URL=${{ secrets.DATABASE_URL }} \
         chatat-server:${{ github.ref_name }} \
         ./chatat-server migrate up
   ```

### Acceptance Criteria:
- [ ] Docker image builds successfully
- [ ] Image pushed to registry on tag
- [ ] Migrations run before deployment
- [ ] Rollback strategy documented

### Testing:
- [ ] Tag v0.1.0 → deploy workflow triggers
- [ ] Docker image runs correctly
- [ ] Migrations complete successfully

---

## Phase 25 Review

### Testing Checklist:
- [ ] Backend CI: lint + test + build
- [ ] Mobile CI: lint + typecheck + test + build
- [ ] Code quality gates: PR checks required
- [ ] Commit lint enforced
- [ ] Docker build works
- [ ] Deploy workflow triggers on tags
- [ ] All workflows pass on main branch

### Review Checklist:
- [ ] CI/CD sesuai `docs/git-workflow.md`
- [ ] Secrets properly configured in GitHub
- [ ] No sensitive data in workflow files
- [ ] Commit: `ci: setup GitHub Actions CI/CD pipeline`
