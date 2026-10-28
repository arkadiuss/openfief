---
title: The read API's use cases compose ledger queries, stored valuations, statuses and FX rates
status: active
sourceFile: internal/views/app/views_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T07:25:30.680Z'
id: TEST-views-app
type: test
---
Tests drive app.Views over in-memory fakes of its ports (Ledger, Valuations, ValuationStatuses, FXRates) holding a synthetic portfolio on 2026-03-02: acc-1 at broker-a in sleeve core with a bond and two ETFs, one of them without a valuation, plus PLN and USD cash; acc-2 at broker-b in sleeve offense with EUR cash. Allocation by every dimension is compared in full with go-cmp (2050.00 PLN in total, cash converted at 4.00 and an older 4.30 rate) and lists the unvalued ETF; a missing FX rate fails with valuation.ErrNoRate and a failing store propagates. AccountValuations and ValuationTotals return stored rows and exact totals per date; ValuationStatus(es) report valued-through, pending and running per account with a never-valued account empty. Flagged assertion filtering, transaction paging, unknown accounts (ErrUnknownAccount on every per-account query) and store failures that must not look like unknown accounts are checked. TestReportingCurrency checks that app.New with PLN reports in PLN and that EUR is refused with a *UnsupportedReportingCurrencyError naming EUR.