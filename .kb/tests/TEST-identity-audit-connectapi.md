---
title: ListAuditEntries over Connect is owner-only, filtered and paged
status: active
sourceFile: internal/identity/connectapi/audit_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-12T19:17:22.475Z'
id: TEST-identity-audit-connectapi
type: test
---
Starts the AuthService on an httptest server with a generated client. TestListAuditEntriesOwnerOnly checks a demo session gets permission denied, and no session, a script token's secret and an unknown session get unauthenticated, as does a caller that is no view plugin. TestListAuditEntries plants an owner session and one entry per audit action, checks no contract value is unspecified, newest first order, paging by token and filters by action, time and actor. TestListAuditEntriesBadRequests checks bad tokens, sizes and enums are invalid arguments. It does not render any page.