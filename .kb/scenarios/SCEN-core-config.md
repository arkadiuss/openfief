---
title: The core reads its connection string from the environment and its settings from flags
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T14:32:54.527Z'
id: SCEN-core-config
type: scenario
---
Given OPENFIEF_DSN and no flags, when the core parses its command line, then it serves with PLN, Europe/Warsaw, end of day 23:30, two workers and the net if redeemed bond basis; given flags, then each overrides its default; given no OPENFIEF_DSN, a market bond basis, an unknown zone or a malformed end of day, then it refuses to start.