---
title: The owner sets a threshold and acknowledges a check through the core
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T02:58:35.898Z'
id: SCEN-valuation-price-check-service
type: scenario
---
Given the core with an owner session and a flagged Sunday mark, when the built-in UI calls PriceCheckService on the plugin port to acknowledge it and to set the instrument's threshold to 5%, then the read API lists it acknowledged and then matched; a threshold of 1 or in percent is invalid, a caller that is not a view plugin is denied, an unknown session is unauthenticated, and an unknown check is not found.