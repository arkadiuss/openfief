---
title: Which accounts each input change revalues, and recording inputs
status: active
sourceFile: internal/valuation/app/plan_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T07:28:27.181Z'
id: TEST-valuation-app-plan
type: test
---
With an in-memory ledger of three synthetic accounts, the Planner revalues for a price only the account that held the instrument from the price's date, for bond terms the holders of the ISIN from their first transaction, for an FX rate or index reading every account with a log, for the bond basis every account from its first transaction, for a tax wrapper only its account, for the end of day each account from its valued-through date but never before its first transaction, and for an import each changed account from its earliest date, coalesced. The InputRecorder hands each change with those revaluations to a fake recorder, rejects invalid prices, rates, readings, terms and wrappers without storing anything, and applies the bond basis, rejecting market.