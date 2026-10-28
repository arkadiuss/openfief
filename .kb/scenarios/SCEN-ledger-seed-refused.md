---
title: Seeding an already seeded account with different content is refused
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:38:40.378Z'
id: SCEN-ledger-seed-refused
type: scenario
---
Given an account whose current transactions include an opening lot or opening cash, when a seeding of it brings a changed or new transaction, then the whole seeding is refused with an *AlreadySeededError naming the account and nothing is appended, for any account in the seeding.