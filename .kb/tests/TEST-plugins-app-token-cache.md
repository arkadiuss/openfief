---
title: Accepted plugin tokens are cached for a bounded time and refused ones never
status: active
sourceFile: internal/plugins/app/tokencache_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T08:46:05.023Z'
id: TEST-plugins-app-token-cache
type: test
verification_scope: unit
---
Table tests drive app.CachingTokenReviewer over a counting fake TokenReviewer with an injected clock and synthetic tokens, some shaped as unsigned JWTs with a chosen exp claim. They check that an accepted token is reviewed once and served from the cache until the TTL, distinct tokens are cached apart, a refused token is reviewed every time, a token expiring before the TTL is cached only until its exp, one expiring after the TTL only for the TTL, and an already expired one not at all. A bounded-cache test fills 1024 entries and checks that a full cache stops caching until stale entries expire. TTL validation refuses zero, negative and token-lifetime TTLs. A false pass would be a cache serving a token past its expiry; the exp cases guard it.