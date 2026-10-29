---
title: Market-data connectors are asked for every held instrument but bonds, and the model's targets
status: active
sourceFile: internal/ingest/app/pricing_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T16:57:07.138Z'
id: TEST-ingest-app-pricing
type: test
---
Use-case test of app.PricingScope over fakes: holdings of two accounts on today's date, one a treasury bond with stored terms, give each non-bond instrument once; the model portfolio's instrument and index targets are added; a failing ledger or model store fails the scope. It asks for holdings on the date Today gives, so a newly bought instrument is in the scope from the sync after its purchase.