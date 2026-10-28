---
title: Per-account queries answer from the ledger and refuse unknown accounts
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T07:25:33.101Z'
id: SCEN-views-app-queries
type: scenario
---
Given the ledger's queries over stored accounts, when the read API asks for an account's transactions, holdings, lots, cash or balance assertion checks, then it gets the ledger's answer, flagged checks alone when it asks for them, and transactions one page at a time; when it names an account that is not stored, then the query fails with an unknown account error, and a store failure is reported as a failure, never as an unknown account.