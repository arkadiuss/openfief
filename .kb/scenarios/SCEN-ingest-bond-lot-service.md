---
title: A view plugin hands the owner's typed lot to the core, which answers with the batch or the reason
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T01:14:00.444Z'
id: SCEN-ingest-bond-lot-service
type: scenario
---
Given a call from a view plugin with the owner's origin and session, when it adds a lot typed with rates in percent and a fee in PLN, then the core reads the rates as fractions, runs the hand-off, and answers the batch's counts, or UNAUTHENTICATED, PERMISSION_DENIED, INVALID_ARGUMENT with the rule, FAILED_PRECONDITION, UNAVAILABLE or INTERNAL without the cause.