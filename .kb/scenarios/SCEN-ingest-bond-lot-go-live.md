---
title: The owner's bond lots go to the connector with the account's go-live
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T06:33:08.531Z'
id: SCEN-ingest-bond-lot-go-live
type: scenario
---
Given a bond account bound to one treasury-bonds instance, when the owner adds or deletes a lot, then the core sends the account's seeding date as go-live, or today when it is not seeded, and imports the delivery through the ledger's path that adds opening lots one at a time, never seeding it again.