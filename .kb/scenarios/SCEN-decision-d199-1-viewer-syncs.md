---
title: A viewer sees only the syncs of accounts it may view
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T06:14:59.249Z'
id: SCEN-decision-d199-1-viewer-syncs
type: scenario
---
Given a viewer granted acc-1 and connector instances bound to acc-1, to acc-2 and to no account, when the viewer reads syncs, sync attempts and the recent activity, then it sees only those of the instance bound to acc-1 and imports into acc-1, and nothing tied only to an instance.