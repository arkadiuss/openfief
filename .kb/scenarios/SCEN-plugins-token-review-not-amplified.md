---
title: No caller can make the core or the egress proxy call TokenReview at will
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T00:10:39.270Z'
id: SCEN-plugins-token-review-not-amplified
type: scenario
---
Given the core and the egress proxy, when a token reaches the core's browser port it is checked only as a session or script token and never sent to TokenReview; on the plugin port and at the proxy a token not shaped like a JWT is refused without a review, and a refused token stays refused for ten seconds without another, while refusals never crowd accepted tokens out of the cache.