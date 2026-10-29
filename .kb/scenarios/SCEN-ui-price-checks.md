---
title: The owner sees flagged price checks next to the flagged balance assertions
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T07:03:03.318Z'
id: SCEN-ui-price-checks
type: scenario
---
Given a core with one flagged price check of a synthetic instrument, when the owner or a read session opens the overview, then the total card says 1 flagged price check(s) and links to /price-checks; when they open that page, then it lists the date, the instrument, the valuation price and the institution's price with their sources, the difference as 2.8232% and the 1% threshold, and changes no price.