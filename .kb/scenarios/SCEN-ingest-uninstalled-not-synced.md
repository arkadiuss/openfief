---
title: An uninstalled instance stops syncing and keeps its state
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T16:50:10.535Z'
id: SCEN-ingest-uninstalled-not-synced
type: scenario
---
Given two connector instances that synced, when one is uninstalled and no longer listed as running, then later ticks sync only the other, and the uninstalled instance's watermark, batches and bindings stay so a reinstall under the same name continues from them.