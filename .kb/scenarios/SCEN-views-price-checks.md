---
title: The read API lists price checks with the threshold, flagged ones on request
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T07:03:01.343Z'
id: SCEN-views-price-checks
type: scenario
---
Given a core that compared a synthetic eodhd close of 10.50 USD with IBKR's marks of 10.52 and 10.8 USD on two days at a 1% threshold, when a client calls GET /v1/price-checks?status=PRICE_CHECK_STATUS_FLAGGED with a valid token, then it gets threshold 0.01 and only the 10.8 USD check, relative difference 0.028571, flagged, with both sources; without a status it gets both; with an unknown status or no token it is refused.