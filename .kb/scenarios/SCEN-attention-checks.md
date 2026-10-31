---
title: Each core check finds its conditions
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T18:31:18.522Z'
id: SCEN-attention-checks
type: scenario
---
Given synthetic records with one flagged assertion, one broken transfer, one flagged price check, one unmapped broker account, one pending ticker, one batch with unsupported rows and an IKZE account without the marginal rate, each check returns exactly that condition, and the core's checks raise the unmapped account's item the read API lists.