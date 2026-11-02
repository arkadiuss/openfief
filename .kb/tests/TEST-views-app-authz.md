---
title: The views compute over the accounts the caller may view and hide the others as unknown
status: active
sourceFile: internal/views/app/access_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T22:45:57.532Z'
id: TEST-views-app-authz
type: test
---
With a fake authorizer that denies acc-2, the account list, valuation totals, statuses, the allocation and performance equal those of the same fixture without acc-2; every single-account query fails for acc-2 exactly as for an account not stored; transfers with a hidden leg are left out; a failing authorizer fails the query; a request without a caller views nothing. The fake records that it was asked about the request's caller.