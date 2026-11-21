---
title: A minute's lookups beyond the limit are refused before the connector is asked
status: active
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md)
  recorded_at: '2026-11-21T12:00:00.000Z'
id: SCEN-views-market-quote-rate-limit
type: scenario
---
Given views whose limit is two lookups a minute, when three are asked in one minute, then the third is resource exhausted and the source was asked twice, and a lookup in the next minute is allowed. Assumes the clock is the views' own.
