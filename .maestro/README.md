# Maestro E2E Test Flows

Catatan Keluarga (ChatAt) end-to-end test flows using [Maestro](https://maestro.mobile.dev/).

## Prerequisites

1. Install Maestro: `curl -Ls "https://get.maestro.mobile.dev" | bash`
2. Start the API server
3. Run the app on a simulator/emulator

## Running Tests

```bash
# Run all flows
maestro test .maestro/flows/

# Run a specific flow
maestro test .maestro/flows/auth-flow.yaml

# Run with video recording
maestro test --format junit .maestro/flows/
```

## Test Accounts

For E2E testing, use the following test accounts:
- Phone: `+6281200000001` (test user 1)
- Phone: `+6281200000002` (test user 2)
- Test OTP code: `123456` (configured in test environment)

## Environment Setup

Set `TEST_MODE=true` in the server environment to enable:
- Predetermined OTP codes (123456)
- Faster token expiry for testing
- Test data seed endpoints
