---
title: An account's payout setting pays out only from the day it took effect and keeps the settings in force before a change
status: active
sourceFile: internal/ingest/payoutsetting_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-03T11:48:59.848Z'
id: TEST-ingest-payout-setting
type: test
---
Table tests of ingest.PayoutSetting: the default keeps payouts as cash, a first paid-out setting covers the whole history, a change takes effect from the day given and replaces later ones, what fell due before follows the earlier setting, the setting keys parse and print as cash and paid-out, and settings out of order or unknown are refused. A false pass would be comparing payouts without dates; each case names the day asked about.