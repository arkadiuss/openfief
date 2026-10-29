---
title: A delivery's instance is recorded as a source of each kind it carries
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T18:00:58.797Z'
id: SCEN-ingest-market-data-sources-recorded
type: scenario
---
Given a delivery from institution xtb by instance xtb-main with a price, an FX rate and an index reading, the core records xtb-main from xtb as a source of prices, FX rates and index readings before storing the values; a delivery without market data records none.