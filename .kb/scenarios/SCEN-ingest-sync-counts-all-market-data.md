---
title: A sync counts every kind of market data it changed and none it re-sent
status: active
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:30:07.177Z'
id: SCEN-ingest-sync-counts-all-market-data
type: scenario
---
Given a connector sync, when it brings FX rates, index readings or index levels that are new or differ from the stored ones, then its entry carries a count of each, and when it brings the same ones again then nothing is written, the counts are zero and no entry is made.

## Context

The stores' upserts skip a row that already holds the incoming values (IS DISTINCT FROM) and count the rows they affected; an index level counts when its source changed too. The entry's counts are fx_rates, index_readings and index_levels beside transactions, prices and bond_terms, and never an amount. Cross-check prices are not counted.

## Source

> connector_synced ALSO COUNTS FX RATES AND INDEX LEVELS (all market data a sync writes), via the same changed-rows upsert (IS DISTINCT FROM, count affected rows) so identical re-sends write nothing.

Source: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18