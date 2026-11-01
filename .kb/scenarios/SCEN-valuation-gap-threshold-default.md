---
title: Gap thresholds by default and the owner's
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T08:55:34.971Z'
id: SCEN-valuation-gap-threshold-default
type: scenario
---
Given no threshold set, prices and FX rates use 7 days, which covers a long weekend with holidays, and index readings 92 days, which covers CPI published in the middle of the month after the one it measures; when the owner sets 14 days for prices it applies to prices only, and 0 or 3661 days is refused.