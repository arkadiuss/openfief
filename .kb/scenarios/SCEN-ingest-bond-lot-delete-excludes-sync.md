---
title: A bond lot deletion and a sync of its instance never overlap
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T11:49:59.410Z'
id: SCEN-ingest-bond-lot-delete-excludes-sync
type: scenario
---
Given a bond account whose lot the treasury-bonds instance books coupons for, when a sync of the instance has read the account's open lots and the owner deletes the lot before the sync's coupons are imported, then the deletion is refused until the sync ends, and deleting it then deletes the coupons the sync booked; and when the deletion runs first, a sync starting meanwhile is refused before it reads the lots, so no coupon of the deleted lot is ever booked (D-154).