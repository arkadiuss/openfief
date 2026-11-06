---
title: Recent activity pages newest first across imports, syncs and the ledger
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T01:58:25.510Z'
id: SCEN-views-activity
type: scenario
---
Given completed imports and syncs, failed sync attempts, a coupon and a correction, when a read caller reads the activity a page at a time, then every entry comes once, newest first, with the page token continuing each source where it stopped, and a coupon or correction in an account the caller may not view is left out.