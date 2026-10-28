---
title: Changing the bond basis revalues every account from its start
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T07:29:03.612Z'
id: SCEN-valuation-basis-change
type: scenario
---
Given two accounts with transactions, when the configured bond basis is applied and differs from the recorded one, then every account is revalued from its first transaction; when it is the same, then nothing is revalued; when market is configured, then applying it fails.