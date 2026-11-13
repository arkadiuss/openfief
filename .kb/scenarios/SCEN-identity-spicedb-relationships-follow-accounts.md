---
title: SpiceDB's account relationships follow the accounts and sleeves
status: active
origin:
  kind: agent
  recorded_at: '2026-11-12T22:51:42.361Z'
id: SCEN-identity-spicedb-relationships-follow-accounts
type: scenario
---
Given a SpiceDB with grants, when an account is placed in a sleeve, moved or taken out of it, or deleted, and when the core starts, then the account's household and sleeve relationships match the ledger and the portfolio, the old sleeve's grant stops covering a moved account, a failed write after the commit is reported and repaired by a retry or the next reconcile, and no grant is added or removed. Assumes writes happen after the database commit.