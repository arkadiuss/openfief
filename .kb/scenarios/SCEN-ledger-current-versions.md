---
title: Only current versions count
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:32:12.984Z'
id: SCEN-ledger-current-versions
type: scenario
---
Given a log with versions 1 and 2 of a buy and one version of two deposits, when the current transactions are taken, then the result holds version 2 of the buy and both deposits, ordered by date and then by key, and version 1 of the buy is left out of every derived holding, lot and balance.