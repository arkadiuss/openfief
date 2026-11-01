---
title: A check's findings raise, change and clear items
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T18:31:05.383Z'
id: SCEN-attention-reconcile
type: scenario
---
Given a stored open item for a broker account, when a check finds that account again with the same detail it changes nothing, when it finds another detail the item changes and stays open, when it no longer finds it the item is resolved as cleared, and when it finds a new subject a new item is raised. Assumes synthetic subjects.