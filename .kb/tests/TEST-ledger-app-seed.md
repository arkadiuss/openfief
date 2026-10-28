---
title: Ledger.Seed refuses a re-seed and serializes concurrent seedings
status: active
sourceFile: internal/ledger/app/seed_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T08:38:15.566Z'
id: TEST-ledger-app-seed
type: test
---
App tests over the in-memory store with per-account locks: a first seeding appends the opening lot and opening cash and checks assertions; the same seeding again reports the account in AlreadySeeded and appends nothing; a changed seeding, or one of two accounts where one is seeded, is refused whole with an *AlreadySeededError and rolls back; an unknown account fails; eight concurrent seedings of one account on different go-live dates leave exactly one.