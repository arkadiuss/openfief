---
title: A payout setting pays out from the day it took effect
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T07:44:16.653Z'
id: SCEN-ledger-payout-setting
type: scenario
---
Given an account paid out since a day, when something falls due on that day or later it is paid out and before it it stays as cash; a paid-out setting without its day, or an unknown one, makes the account invalid; the setting and its day are stored with the account.