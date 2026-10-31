---
title: The sync creates every owned account, cash accounts included, and touches nothing else
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T19:00:16.120Z'
id: SCEN-views-ghostfolio-owned-and-cash-accounts-created
type: scenario
---
Given the projected owned accounts, the record's account's own and one per other currency of its cash, when a sync runs against Ghostfolio, then the accounts missing there are created with their derived UUIDs and currencies, the owned accounts' activities and balances are read and reconciled, an activity in a cash account is stale and deleted, and every account and activity outside the owned accounts is left as it was.