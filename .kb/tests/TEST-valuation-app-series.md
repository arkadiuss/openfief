---
title: The owner's series sources and gap thresholds, and gap items raised by a revaluation
status: active
sourceFile: internal/valuation/app/series_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T08:55:08.597Z'
id: TEST-valuation-app-series
type: test
---
Use-case tests over in-memory fakes: SetSeriesSource lets the owner pick only a source with a value of the series, revalues from their start only the accounts that held the instrument (every account for an FX rate), revalues nothing for the same source again, and refuses another session or an invalid series; SetGapThreshold stores the owner's threshold, which SourceRankings returns, and refuses zero days or another session; Revalue over a synthetic ETF raises a market-data-gap item for the account and the price series when the last price is 9 days old, none under the owner's 10-day threshold, and clears the item when prices resume, clearing the FX series that has no gap; NeededSeries picks prices, FX rates and the bond type's period index for the holdings valued on the day only.