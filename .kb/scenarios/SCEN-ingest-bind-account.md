---
title: The owner binds a connector instance to a stored account
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T16:50:06.507Z'
id: SCEN-ingest-bind-account
type: scenario
---
Given a stored ledger account acc-1, when the owner runs core connector bind for instance xtb-main and acc-1, then the binding is stored once however often it is repeated; when the account is not stored or the instance name is not a plugin name, then nothing is stored and the command fails.