---
title: An unmapped broker account's outbox event names only its instance
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T15:54:50.166Z'
id: SCEN-attention-outbox-read-subject
type: scenario
---
Given an attention item about an unmapped broker account, whose subject is the connector instance and the broker account number, when its change is written with its outbox event, then the event's subject is the one read access sees, the instance alone, and the item keeps its full subject for the owner (D-121, D-165).