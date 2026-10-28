---
title: Sync states, batches and bindings survive in the ingest schema
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T16:50:16.897Z'
id: SCEN-ingest-sync-store-roundtrip
type: scenario
---
Given a migrated ingest schema, when an instance is claimed, a batch started, failed and completed, then the claim, the open batch with its mode, the failure count, the watermark, the seeded flag and the completed batch's counts load back; when a claim's lease has expired, then a new claim succeeds; when a row has an invalid instance name, then PostgreSQL refuses it.