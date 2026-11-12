---
title: A second Ghostfolio sync changes nothing
status: active
tags:
  - acceptance:p10-9
origin:
  kind: agent
  recorded_at: '2026-11-08T08:59:31.641Z'
id: SCEN-e2e-ghostfolio-second-sync-unchanged
type: scenario
---
Given the view's first sync has made the owned accounts hold the projection, when the view syncs again with the record unchanged, then the sync reports no account created, no activity imported or deleted, no account cleared and no balance set or deleted, and every activity and balance of the owned accounts is still there with the same Ghostfolio ID and fields. Runs on both kind and k3d.