---
title: SpiceDB with the household grants answers as the policy does
status: active
origin:
  kind: agent
  recorded_at: '2026-11-12T22:51:42.252Z'
id: SCEN-identity-spicedb-matches-policy
type: scenario
---
Given a SpiceDB holding authz/schema.zed, the accounts and sleeves of the ledger and portfolio and a household grant for each script token and view plugin that reads today, when the owner, a demo session, a script token, a view plugin or nobody asks to view or change each account, then SpiceDB answers exactly as identity.MayAct does, and lists the accounts as it would one by one. Assumes the demo sessions' wildcard grant is written because demo sessions exist.