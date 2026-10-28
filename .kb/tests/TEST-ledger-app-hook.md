---
title: Import hooks run inside the import's transaction and roll it back on failure
status: active
sourceFile: internal/ledger/app/hook_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:28:46.182Z'
id: TEST-ledger-app-hook
type: test
---
An import of one synthetic deposit with a matching cash assertion runs over the in-memory store with a hook that first checks, through the transaction's Store, that it sees the appended version. When the hook succeeds it receives the summary with New 1, the earliest affected date 2026-03-02 for acc-1 and the matched check, and the version and check are stored. When the hook fails with a stand-in enqueue error, Import returns that error (errors.Is), and neither the version nor the check is stored, so the import rolled back whole.