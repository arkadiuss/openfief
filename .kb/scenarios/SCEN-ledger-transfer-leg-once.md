---
title: A leg belongs to at most one confirmed transfer
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T06:05:38.390Z'
id: SCEN-ledger-transfer-leg-once
type: scenario
---
Given two withdrawals of 100 PLN and one deposit of 100 PLN into another account, when the owner confirms the first withdrawal with the deposit, confirming the second with it is refused as taken; when both confirmations arrive at once, exactly one is stored. Rejecting a pair whose leg is taken is allowed.