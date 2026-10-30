---
title: A manually entered account receives opening lots one at a time on its go-live
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T05:01:40.495Z'
id: SCEN-ledger-opening-lots-one-at-a-time
type: scenario
---
Given an account seeded with an opening lot on its go-live date, when another opening lot dated that day is added, then it is imported and the same lot again changes nothing; a lot dated another day, opening cash, or a first lot into an unseeded account that already has a transaction on or before its date is refused.