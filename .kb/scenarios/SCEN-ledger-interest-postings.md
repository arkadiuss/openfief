---
title: Interest is booked like a dividend under its own kind
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T05:01:43.537Z'
id: SCEN-ledger-interest-postings
type: scenario
---
Given a bond coupon of a gross amount with Belka tax withheld, when it is booked as interest, then the holding posts minus the gross with no units, withholding tax the withheld part and cash the net, under the kind interest rather than dividend.