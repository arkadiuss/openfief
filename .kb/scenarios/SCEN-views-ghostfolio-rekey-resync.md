---
title: Ghostfolio resyncs a re-keyed bond without duplicates
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T22:44:32.767Z'
id: SCEN-views-ghostfolio-rekey-resync
type: scenario
---
Given the Ghostfolio view synced a bond held under its provisional ISIN, when the core re-keys it to the registry's ISIN and the view syncs again, then nothing is imported or deleted and the activities stay as they were.