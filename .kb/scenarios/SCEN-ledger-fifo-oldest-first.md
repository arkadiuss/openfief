---
title: A sale consumes the oldest lot first
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:39:03.761Z'
id: SCEN-ledger-fifo-oldest-first
type: scenario
---
Given an opening lot of 3 units opened in 2019 and two later buys on the same day, when 5 units are sold, then the opening lot is consumed whole and 2 units come from the buy whose key sorts first, and the other buy is untouched; on one day, buys are applied before sales, so a sale whose row comes first is still covered.