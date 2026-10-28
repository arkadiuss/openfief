---
title: The demo dataset holds no real account data
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T09:30:38.758Z'
id: SCEN-demo-synthetic-only
type: scenario
---
Given the generated dataset, when its instruments, bonds, Flex statements and ledger transactions are inspected, then every ISIN is valid and carries the TESTD marker after its country code, every IBKR statement names only the placeholder account U0000000, every account ID starts with demo-, and the only real figures are the public bond terms, GUS CPI readings and NBP rates.