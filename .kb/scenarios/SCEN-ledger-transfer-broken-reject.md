---
title: A broken confirmed transfer can be rejected
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T18:31:26.753Z'
id: SCEN-ledger-transfer-broken-reject
type: scenario
---
Given a confirmed pair whose deposit a correction changed, confirming it again is refused, rejecting it is stored as the pair's next version, and a pair never confirmed is still checked when rejected.