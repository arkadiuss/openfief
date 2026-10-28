---
title: Concurrent seedings of one account seed it once
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:38:51.130Z'
id: SCEN-ledger-seed-concurrent
type: scenario
---
Given an unseeded account, when several seedings of it with different keys run at the same time, then exactly one commits and every other one is refused with an *AlreadySeededError, because each seeding locks the account in its database transaction before it checks and appends.