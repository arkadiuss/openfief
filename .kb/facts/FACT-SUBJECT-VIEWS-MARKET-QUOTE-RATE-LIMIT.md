---
title: Views Market Quote Rate Limit
status: active
fact_kind: subject
subject_key: views.market_quote_rate_limit
canonical_key: views.market_quote_rate_limit
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md)
  recorded_at: '2026-11-21T12:00:00.000Z'
id: FACT-SUBJECT-VIEWS-MARKET-QUOTE-RATE-LIMIT
type: fact
---
How many market quote lookups the views make in a minute, all callers together, which spares the owner's provider key; it is not a limit on the read API's other calls.
