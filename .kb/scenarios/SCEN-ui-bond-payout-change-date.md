---
title: The account form takes the day a payout change takes effect from
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T16:48:09.703Z'
id: SCEN-ui-bond-payout-change-date
type: scenario
---
Given a hand-entered bond account's change form, when the owner opens it, then it offers the payout setting and a date field for the day the change takes effect from, and the posted day reaches the core with the setting, empty meaning today; other accounts' forms offer neither.