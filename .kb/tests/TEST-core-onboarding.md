---
title: The real core takes the owner's sleeves and accounts and books a mapped broker account
status: active
sourceFile: cmd/core/onboarding_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T12:45:24.379Z'
id: TEST-core-onboarding
type: test
verification_scope: integration
---
PostgreSQL test through the real core (task test:pg) after a passkey login: sleeves and kinds through the SleeveService, an IKZE account with a USD-minimum fee rule for the broker account U0000001 through the AccountService, answered with its terms, a second mapping refused, and a delivery booked into the mapped account.