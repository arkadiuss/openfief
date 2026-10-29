---
title: The real core takes the owner's sleeves and accounts and books a mapped broker account
status: active
sourceFile: cmd/core/onboarding_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T12:45:24.379Z'
id: TEST-core-onboarding
type: test
---
PostgreSQL test under task test:pg: the core starts with the default sleeves, the owner logs in with a virtual passkey, adds a kind and a sleeve through SleeveService and is refused a role other than the kind's; a broker connector's sync books nothing for its two unmapped synthetic broker accounts, which the read API lists; the owner creates the account for one through AccountService; the next sync books the deposit there with the account in its key, and the read API lists the binding and only the other as unmapped.