---
title: The docs build ingests both clusters' proof runs, oldest first, and stops at a refusal
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T12:45:35.294Z'
id: SCEN-tooling-kibi-ingest-oldest-first
type: scenario
---
Given the proof-run artifacts of the k3d and kind end-to-end jobs downloaded under one directory, when task kibi:ingest runs, then kibi ingest-proof receives each artifact whole with its own snapshot, in the order the runs finished, so each test's receipts stay chronological; with no artifact, or when kibi refuses one, the build fails and nothing later is ingested.