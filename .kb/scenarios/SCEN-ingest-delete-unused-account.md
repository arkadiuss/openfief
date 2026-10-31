---
title: An account with no transactions is deleted; one with transactions stays
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T14:35:09.925Z'
id: SCEN-ingest-delete-unused-account
type: scenario
---
Given an account that holds a transaction and one that holds none, when the owner deletes each, then the first is refused and nothing is changed, and the second goes with its bindings, broker account mappings, tax wrapper, fee rule and stored valuations; deleting it again changes nothing.