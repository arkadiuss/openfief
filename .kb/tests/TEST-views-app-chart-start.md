---
title: The read API's performance starts on the stored charts' start, and only unmarked seeded accounts move the default
status: active
sourceFile: internal/views/app/chartstart_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-02T08:22:13.512Z'
id: TEST-views-app-chart-start
type: test
---
Runs the views' Performance, AccountPerformance and ModelPortfolio over in-memory fakes of the ledger, valuations and the stored charts' starts. A stored portfolio start is used when the query names none and named as Chosen, a requested start previews over it, an account uses its own stored start, and the model's line starts on the stored start. A seeded account marked as opened on its seeding day leaves the default start where it was, an unmarked one moves it. An account's stored start is read only after the authorizer let the caller view the account, and a store failure fails the query instead of falling back silently.