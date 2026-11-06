---
title: The Inbox views, the activity, the sync history and the instrument page over in-memory ports
status: active
sourceFile: internal/views/app/uiv2_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T01:58:21.913Z'
id: TEST-views-app-uiv2
type: test
---
Runs views/app over fakes of the ledger, attention items, ingest history, ledger changes and prices with synthetic accounts and ISINs. Checks attention views, filters and counts, each item's scope and group, an item of an account the caller may not view left out; the activity paged two at a time with a hidden account's correction left out; sync statuses and attempts paged, an unknown instance refused; an instrument's listings, sources, classes and prices; its transactions and holdings across viewable accounts only; its price checks.