---
title: One account's performance starts on its own first day and counts its deposits as cash flows
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T13:01:28.946Z'
id: SCEN-views-account-performance
type: scenario
---
Given the same synthetic accounts, when one account's performance is asked for, then its index is 100 on its own first day, whatever day another account was seeded, a chosen start rebases it there, a seeded account starts on its seeding day with its own pending revaluation, an account with no transactions has no days and no start, and an unknown account is refused as unknown.