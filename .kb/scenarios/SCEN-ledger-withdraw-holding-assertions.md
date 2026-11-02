---
title: The ledger withdraws the holding assertions a correction replaces
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T07:00:26.924Z'
id: SCEN-ledger-withdraw-holding-assertions
type: scenario
---
Given a holding assertion of the instrument a ticker's ISIN was corrected from, when it is withdrawn, then its check is removed and withdrawing it again changes nothing, while a cash assertion is refused and nothing is removed; the transaction log is untouched.