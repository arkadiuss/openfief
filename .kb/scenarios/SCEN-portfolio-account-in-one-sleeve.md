---
title: An account the ledger holds is put in one stored sleeve, and leaves it when deleted
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T11:49:19.463Z'
id: SCEN-portfolio-account-in-one-sleeve
type: scenario
---
Given sleeves core and play and accounts acc-1 and acc-2 in the ledger, when acc-2 is placed in core and then in play, then it is in play only; placing an account the ledger doesn't hold or in a sleeve that isn't stored is refused; and when an account is deleted it leaves its sleeve, which can then be deleted.