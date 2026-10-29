---
title: The core hands an owner's bond lot to the account's one bound connector and imports what it delivers
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T01:13:56.462Z'
id: SCEN-ingest-bond-lot-handoff
type: scenario
---
Given an owner-scoped session and an account bound to exactly one connector instance, when the owner adds a valid lot, then the core hands it to that instance as a new transactions batch with source entry, imports the delivered transactions with keys scoped to the instance and records the delivered series terms; a missing or read-scoped session, an invalid lot, zero or several bound instances, a refusal and a delivery into another account import nothing.