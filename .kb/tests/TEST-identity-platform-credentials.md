---
title: A platform credential is 32 or more random bytes as lowercase hex, and the platform's Secrets are exactly SpiceDB's key and its role's basic-auth Secret
status: active
sourceFile: internal/identity/platform_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-13T07:41:43.681Z'
id: TEST-identity-platform-credentials
type: test
---
Table tests of identity.PlatformCredential and identity.PlatformSecrets, pure functions. 32 bytes become 64 lowercase hex digits and more bytes keep every one; 31 bytes and none are refused with ErrShortCredential. PlatformSecrets returns, in order, openfief-spicedb-key (type Opaque, one key) and openfief-db-spicedb (type kubernetes.io/basic-auth, username spicedb and a password, with the cnpg.io/reload label), each with the app.kubernetes.io labels, holding the two values given. A false pass would be a function that returned a third Secret or a differently named one, which the whole-list comparison rules out.