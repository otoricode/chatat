#!/usr/bin/env bash
# scripts/test-report.sh - Generate comprehensive test report for Chatat
#
# Usage: ./scripts/test-report.sh [--ci]
# Outputs a Markdown test report to stdout and optionally saves to tmp/test-report.md

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
CI_MODE="${1:-}"
REPORT_FILE="$ROOT_DIR/tmp/test-report.md"

mkdir -p "$ROOT_DIR/tmp"

# Colors for terminal output (only in non-CI mode)
if [ "$CI_MODE" = "--ci" ]; then
  RED=""
  GREEN=""
  YELLOW=""
  NC=""
else
  RED='\033[0;31m'
  GREEN='\033[0;32m'
  YELLOW='\033[1;33m'
  NC='\033[0m'
fi

echo -e "${GREEN}Generating test report...${NC}"

# ─── Go Backend Tests ─────────────────────────────────────────
echo -e "${YELLOW}Running Go tests with coverage...${NC}"
cd "$ROOT_DIR/server"

GO_TEST_OUTPUT=$(go test -short -count=1 -cover ./internal/... ./pkg/... 2>&1 || true)
GO_PASS_COUNT=$(echo "$GO_TEST_OUTPUT" | grep -c "^ok" || true)
GO_FAIL_COUNT=$(echo "$GO_TEST_OUTPUT" | grep -c "^FAIL" || true)

# Extract coverage per package (package name + percentage)
GO_COVERAGE=$(echo "$GO_TEST_OUTPUT" | grep -E "^ok.*coverage:" | awk '{for(i=1;i<=NF;i++) if($i ~ /%/) pct=$i; gsub(/github.com\/otoritech\/chatat\//, "", $2); print $2, pct}')

# ─── Mobile Tests ─────────────────────────────────────────────
echo -e "${YELLOW}Running mobile tests with coverage...${NC}"
cd "$ROOT_DIR/mobile"

MOBILE_TEST_OUTPUT=$(npx jest --no-coverage --silent 2>&1 || true)
MOBILE_SUITES=$(echo "$MOBILE_TEST_OUTPUT" | grep "Test Suites:" | sed 's/.*[^0-9]\([0-9]*\) passed.*/\1/' | head -1 || echo "0")
MOBILE_TESTS=$(echo "$MOBILE_TEST_OUTPUT" | grep "^Tests:" | sed 's/.*[^0-9]\([0-9]*\) passed.*/\1/' | head -1 || echo "0")
MOBILE_PASS=$(echo "$MOBILE_TEST_OUTPUT" | grep "passed" | tail -1 || echo "")

# ─── Generate Report ──────────────────────────────────────────
cd "$ROOT_DIR"

TIMESTAMP=$(date -u '+%Y-%m-%d %H:%M:%S UTC')
GIT_SHA=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH=$(git branch --show-current 2>/dev/null || echo "unknown")

cat > "$REPORT_FILE" << HEREDOC
# Chatat Test Report

**Generated:** $TIMESTAMP
**Branch:** $GIT_BRANCH
**Commit:** $GIT_SHA

---

## Summary

| Component | Suites | Status |
|-----------|--------|--------|
| Go Backend | $GO_PASS_COUNT passed, $GO_FAIL_COUNT failed | $([ "$GO_FAIL_COUNT" = "0" ] && echo "PASS" || echo "FAIL") |
| Mobile (RN) | $MOBILE_SUITES suites | $MOBILE_PASS |
| E2E (Maestro) | 4 flows | Ready (manual run) |

---

## Go Backend Coverage

| Package | Coverage |
|---------|----------|
$(echo "$GO_COVERAGE" | while read -r pkg cov; do echo "| $pkg | $cov |"; done)

---

## Go Test Output

\`\`\`
$GO_TEST_OUTPUT
\`\`\`

---

## Mobile Test Output

\`\`\`
$MOBILE_TEST_OUTPUT
\`\`\`

---

## E2E Test Flows

| Flow | File | Status |
|------|------|--------|
| Auth | .maestro/flows/auth-flow.yaml | Ready |
| Send Message | .maestro/flows/send-message.yaml | Ready |
| Create Document | .maestro/flows/create-document.yaml | Ready |
| Create Group | .maestro/flows/create-group.yaml | Ready |

> Run E2E tests: \`maestro test .maestro/flows/\`
HEREDOC

echo -e "${GREEN}Report saved to: $REPORT_FILE${NC}"

# Print summary to stdout
echo ""
echo "=== Test Report Summary ==="
echo "Go: $GO_PASS_COUNT passed, $GO_FAIL_COUNT failed"
echo "Mobile: $MOBILE_PASS"
echo "E2E: 4 Maestro flows (ready)"
echo "Full report: $REPORT_FILE"
