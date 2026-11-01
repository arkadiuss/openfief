---
title: A payout setting keeps the settings in force before a change
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T16:47:59.455Z'
id: SCEN-ledger-payout-setting-history
type: scenario
---
Given a bond account whose first payout setting covers its whole history, when the owner changes it from a day, possibly earlier than today, then what falls due from that day follows the new setting, what fell due before follows the setting in force then, settings from that day on are replaced, and the setting with its earlier ones is stored and read back unchanged.