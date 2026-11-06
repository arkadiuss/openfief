---
title: A batch as a caller who may view some of its accounts sees it
status: active
sourceFile: internal/views/scope_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T11:51:11.428Z'
id: TEST-views-scope-batch
type: test
---
Table test of views.ScopeBatch: every account viewable keeps the totals; one account viewable sums its counts alone and drops the rows tied to no account; none viewable names none; a batch without counts per account shows none to a partial view; a batch that booked into no account is unchanged.