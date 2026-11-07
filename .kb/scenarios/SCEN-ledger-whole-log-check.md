---
title: The whole-log check names unbalanced versions and counts superseded ones
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T04:21:40.567Z'
id: SCEN-ledger-whole-log-check
type: scenario
---
Given stored versions of transactions in several accounts, some superseded, one listed twice and one not balanced, when the whole log is checked, then the counts per account and in all count each version once, each unbalanced version is named per currency with its sum, the per-currency sums of the whole log are given, and core check ledger writes the report as JSON and logs only its counts.