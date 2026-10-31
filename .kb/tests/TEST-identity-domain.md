---
title: Script tokens are validated, expire, revoke and parse as specified
status: active
sourceFile: internal/identity/token_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T14:11:01.603Z'
id: TEST-identity-domain
type: test
verification_scope: unit
---
Table tests of the pure identity package with fixed times: NewScriptToken accepts a valid ID, name, read scope and a lifetime up to MaxTTL and refuses an empty, long or non-ASCII name, a zero, negative or over-a-year lifetime, an unknown scope and a bad ID; Check refuses a revoked token before an expired one, accepts a nanosecond before expiry and refuses at expiry, and refuses a scope not granted; FormatSecret writes the exact base64url text and ParseSecret refuses missing prefixes, short, padded or non-base64url keys without echoing the secret; KeyMatches compares SHA-256 hashes and refuses other keys and truncated hashes. A false pass would be a hash equal to the key, which the test rules out.