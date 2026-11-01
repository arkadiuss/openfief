---
title: Allocation groups valued holdings and cash by asset class, currency, institution or sleeve
status: active
sourceFile: internal/views/allocation_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T07:17:55.472Z'
id: TEST-views-allocation
type: test
verification_scope: unit
---
Table tests run views.Allocate over five synthetic items (a USD ETF, USD and EUR cash, a PLN stock, a treasury bond) for each of the four dimensions and compare the whole Allocation with go-cmp: group keys, exact PLN sums, largest group first, and a total of 2330.50 PLN. Edge cases: no items gives zero PLN and no groups, equal groups sort by key, a PLN value in USD fails with ErrNotPLN and an unknown dimension with ErrUnknownDimension. ParseDimension round-trips every name and rejects broker. AssetClassOf calls a Polish ISIN with a bond series symbol a treasury bond and anything else listed. ValueHoldings pairs holdings with valuations of the same account, instrument and date and returns the one without a valuation as unvalued, ignoring valuations of another date or account.