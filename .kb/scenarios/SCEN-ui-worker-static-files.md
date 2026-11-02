---
title: The service worker keeps only static files; offline it shows the static page, with the kept overview on /
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T10:10:56.895Z'
id: SCEN-ui-worker-static-files
type: scenario
---
Given the service worker installed, when the owner opens pages, loads fragments or posts forms, then none of those responses is stored; only the listed styles, scripts, fonts, icons, the manifest and the offline page are. When the network is gone a navigation shows the offline page with Kasztan asleep; on the overview's address it also shows the last overview offline.js keeps (D-151), and elsewhere nothing of the record. When the owner signs out the cache is emptied and the kept overview deleted, and when a new version ships the old cache is dropped.