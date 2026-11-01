---
title: No go-live goes with an unseeded account's lots
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T07:44:36.403Z'
id: SCEN-ingest-bond-lot-no-go-live
type: scenario
---
Given an account never seeded with opening lots, when the owner enters or deletes a bond lot, then the core hands it to the connector without a go-live date; an account seeded with opening lots gets its seeding date.