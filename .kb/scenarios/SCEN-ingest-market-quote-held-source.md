---
title: A held instrument's quote is the price its valuation uses
status: active
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md), and the design session's ruling on it
  recorded_at: '2026-11-21T14:00:00.000Z'
id: SCEN-ingest-market-quote-held-source
type: scenario
---
Given an owner who holds an instrument whose prices a source other than the first ranked supplies, when its quote is looked up, then that source alone is asked and the quote's latest close is its latest price, so the model never sees two prices of one instrument; an instrument the owner does not hold goes to the ranked sources, and a held one whose source is not running has no source.
