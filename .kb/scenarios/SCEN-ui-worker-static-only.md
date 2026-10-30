---
title: The service worker keeps only static files and shows a static page offline
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T08:07:57.100Z'
id: SCEN-ui-worker-static-only
type: scenario
---
Given the service worker installed, when the owner opens pages, loads fragments or posts forms, then none of those responses is stored; only the listed styles, scripts, fonts, icons, the manifest and the offline page are. When the network is gone a navigation shows the offline page with Kasztan asleep and no data. When the owner signs out the cache is emptied, and when a new version ships the old cache is dropped.