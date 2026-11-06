---
title: A connector instance's sync history is served without error text
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T01:58:26.838Z'
id: SCEN-views-sync-history
type: scenario
---
Given an instance whose latest attempts failed, when a read caller asks for the sync statuses and the instance's attempts, then the status says how many failed since when, the latest error class and whether it keeps failing, and the attempts come newest first a page at a time with their classes; an instance never synced is not found.