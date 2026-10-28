---
title: A posting must fit its ledger account
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:23:58.662Z'
id: SCEN-ledger-posting-ledger-account
type: scenario
---
Given a transaction, when a posting puts USD into the PLN cash ledger account, puts a quantity on cash, or goes to a holding ledger account that also has a currency, then the transaction is rejected as invalid; when a holding posting carries units and the USD value exchanged for them, then it is accepted.