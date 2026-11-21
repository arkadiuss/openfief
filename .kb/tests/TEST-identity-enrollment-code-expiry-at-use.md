---
title: An enrollment code that expires mid-ceremony is refused when the registration finishes, in the store and in the use case
status: active
sourceFile: internal/identity/postgres/invites_test.go
verification_scope: integration
origin:
  kind: agent
  ref: 'Security review before the alpha.2 cut, known limit #4 (2026-11-21)'
  recorded_at: '2026-11-21T18:38:59.016Z'
id: TEST-identity-enrollment-code-expiry-at-use
type: test
---
Against PostgreSQL: UseEnrollmentCode at and after a code's expiry fails with ErrEnrollmentCodeExpired and leaves it unused, a used code still fails as used, and a code a second before its expiry is spent. In the use case: a viewer's invitation that expires between BeginRegistration and FinishRegistration is refused as forbidden and creates no viewer.