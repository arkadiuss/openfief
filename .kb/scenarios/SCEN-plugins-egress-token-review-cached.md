---
title: An accepted token is trusted for a bounded time
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:46:21.866Z'
id: SCEN-plugins-egress-token-review-cached
type: scenario
---
Given a TTL of five minutes and a token the TokenReviewer accepts, when the same token is presented again before five minutes pass, then no second review happens; at five minutes, or at the token's own exp if that is earlier, it is reviewed again; a refused token is reviewed on every request.