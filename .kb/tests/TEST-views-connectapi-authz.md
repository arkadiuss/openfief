---
title: Over REST, the read API lists only viewable accounts and answers a hidden one as missing
status: active
sourceFile: internal/views/connectapi/authz_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-01T22:45:59.646Z'
id: TEST-views-connectapi-authz
type: test
---
Through the real Vanguard transcoder and auth interceptor, an authorizer that lets only the token's caller view acc-1 makes GET /v1/accounts list acc-1 alone, and every account path of acc-2 answer 404 with the same body as an account not stored, while acc-1's paths answer 200.