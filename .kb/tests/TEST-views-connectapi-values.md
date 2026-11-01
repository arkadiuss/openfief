---
title: The read API serves settings, account terms, three values with cash, yearly taxes and allocations of a kind over REST
status: active
sourceFile: internal/views/connectapi/values_test.go
origin:
  kind: agent
  recorded_at: '2026-10-30T00:04:48.363Z'
id: TEST-views-connectapi-values
type: test
verification_scope: unit
---
REST requests through the real Vanguard transcoder over fakes with acc-1 in IKZE: GET /v1/settings, the account's wrapper and fee rule, valuations with net values and the marginal_rate reason, totals of 809.21 gross and 795 net with -360 of cash and no liquidation total, OKI's yearly tax not in force, and a liquidation allocation leaving out the ETF and the cash; TestREST checks each answer against the generated OpenAPI document; holdings and lots read with ibkr-old removed mark the ETF holding and its lot, not the bond, and match the contract.