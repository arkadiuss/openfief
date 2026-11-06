---
title: Each attention kind has its Inbox group, and each subject names its account, instrument or instance
status: active
sourceFile: internal/attention/scope_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T01:26:32.924Z'
id: TEST-attention-scope
type: test
---
Table tests of attention.GroupOf over every kind the core raises, those of the valuation and ingest contexts named by their own constants, and an unknown kind; of Group.String; and of attention.ScopeOf over a holding's and a cash balance assertion, a price check, an instrument to classify, a price and an FX series gap, an unmapped broker account, a ticker, a failing sync, undeclared deliveries, import problems with and without an instance in the detail, a setting, a broken transfer and an unknown kind. An unmapped broker account's scope must not carry the broker account number.