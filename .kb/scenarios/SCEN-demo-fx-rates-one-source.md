---
title: The demo's loaded rates and its NBP instance are one series
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T08:56:29.046Z'
id: SCEN-demo-fx-rates-one-source
type: scenario
---
Given the loader with its default settings, when it loads NBP table A rates, offline or fetched, then every rate names the source nbp, the demo's NBP connector instance, so its later syncs extend the same series instead of becoming cross-checks of it; an empty source is refused.