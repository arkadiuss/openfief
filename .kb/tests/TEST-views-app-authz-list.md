---
title: The views list the viewable accounts in one call when the authorizer can
status: active
sourceFile: internal/views/app/access_list_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-12T22:51:24.933Z'
id: TEST-views-app-authz-list
type: test
---
With a fake authorizer that also lists the accounts a caller may view: Accounts equals the result of asking account by account, with one list call and no check; a failing list leaves an error and no accounts. A false pass would be a fallback to every account on failure, which the failing-list test would catch.