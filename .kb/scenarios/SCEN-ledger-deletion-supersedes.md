---
title: A deleted transaction is superseded and kept
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T05:01:37.859Z'
id: SCEN-ledger-deletion-supersedes
type: scenario
---
Given a transaction recorded under a key, when a deletion of that key is imported, then it is stored as the next version, the transaction no longer counts toward holdings, lots or cash, and its earlier version stays in the log; deleting again changes nothing, importing it again brings it back, and deleting a key never recorded is refused.