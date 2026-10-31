---
title: Script tokens round-trip through PostgreSQL with only a hash stored
status: active
sourceFile: internal/identity/postgres/store_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T14:11:05.960Z'
id: TEST-identity-postgres
type: test
verification_scope: integration
---
Runs under task test:pg on a fresh database migrated twice. Issuing through the use cases stores a token that loads back equal, with the key's SHA-256 hash, and no column holds the key; a second token with the same random ID is refused; the secret verifies, then after two revocations a minute apart verification fails as revoked and listing shows the first revocation time. Unknown IDs fail with ErrUnknownToken for loading and revoking, and the table refuses a short hash, an expiry not after creation, an upper-case ID and a name with a space.