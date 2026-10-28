---
title: An import and the revaluation it enqueues commit together
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:29:04.441Z'
id: SCEN-ledger-app-import-hook
type: scenario
---
Given an import hook that enqueues a revaluation per account from the earliest affected date, when a deposit dated 2026-03-02 is imported into acc-1, then the hook runs in the import's database transaction, sees the appended deposit, and its job commits with the import; when enqueueing fails, then the import is rolled back whole: no version, check or job is stored. A correction that moves a transaction to another account and an earlier date reports both accounts with their own earliest dates.