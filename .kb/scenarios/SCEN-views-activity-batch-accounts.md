---
title: An upload in the recent activity names its viewable accounts
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T03:42:56.800Z'
id: SCEN-views-activity-batch-accounts
type: scenario
---
Given an upload booked into acc-1 and acc-2 and another booked into acc-1 only, when a caller who may not view acc-1 reads the recent activity, then the first upload names only acc-2 and the second is not listed; over Connect the upload's batch carries its account IDs.