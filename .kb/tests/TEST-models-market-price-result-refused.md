---
title: A conversation holding a T2 price goes only to a model allowed T2, at the router and at the Ask service
status: active
sourceFile: internal/modelrouter/app/router_test.go
verification_scope: unit
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md)
  recorded_at: '2026-11-21T12:00:00.000Z'
id: TEST-models-market-price-result-refused
type: test
---
With one model allowed T0 and another T2 and none both, a conversation of a T0 question and a T2 tool result is refused with ErrDataTierDenied and audited as refused, and no provider is called; allowing T2 for the first sends it there. Through the Ask service, the tier check says no for a route allowed T0 only and a generate call is refused as permission denied, with the check and the call both allowing it once T2 is allowed. A false pass would be a fake provider that is called before the tiers are checked.
