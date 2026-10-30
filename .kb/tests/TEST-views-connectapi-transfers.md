---
title: GET /v1/transfers answers suggested and confirmed pairs with their keys, as the OpenAPI document says
status: active
sourceFile: internal/views/connectapi/transfers_test.go
origin:
  kind: agent
  recorded_at: '2026-10-30T06:20:04.203Z'
id: TEST-views-connectapi-transfers
type: test
---
Serves the read API over a fake ledger with two withdrawal and deposit pairs of 250 PLN, one confirmed, and calls GET /v1/transfers with a valid token: the response validates against the generated OpenAPI document and lists the confirmed pair under confirmed and the other under suggested, with accounts, dates, amount and both legs' transaction keys.