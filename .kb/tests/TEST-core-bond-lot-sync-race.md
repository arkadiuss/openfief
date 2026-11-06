---
title: A bond lot deletion and a sync of its instance never overlap, through the real core and PostgreSQL
status: active
verification_scope: integration
sourceFile: cmd/core/couponrace_test.go
origin:
  kind: agent
  recorded_at: '2026-11-02T11:49:56.673Z'
id: TEST-core-bond-lot-sync-race
type: test
---
Reproduces D-154's race both ways through the real core over a fresh PostgreSQL database and a connector served over HTTP whose sync and bond lot import can be held at a gate. A sync that has read the bond account's open lot is held while the owner deletes the lot: the deletion must be refused with ErrLotSyncing, and once the sync has booked its two coupons a second deletion must supersede the lot and both coupons, leaving no current transaction. A deletion held inside its connector call must make a sync fail with ErrSyncRunning before the connector is called, and the sync after the deletion must book nothing. Without the claim both tests fail: the coupons of the deleted lot stay booked.