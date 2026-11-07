---
title: openfief check balances prints to the terminal and exits non-zero on a failure
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T23:37:12.892Z'
id: SCEN-views-cli-check-balances
type: scenario
---
Given a core whose database holds a synthetic account with a matching cash assertion and a holding assertion a unit off, when the owner runs openfief check balances, then the core writes the report as JSON to the exec's output and logs only that the check ran with its counts; the CLI prints a table, or the JSON with --json, and exits with status 1 after printing; arguments it does not take are a usage error before the cluster is reached.