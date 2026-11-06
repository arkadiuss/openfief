---
title: Coupons and corrections page newest first
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T01:12:20.050Z'
id: SCEN-ledger-recorded-changes-pages
type: scenario
---
Given a ledger log with first versions, corrections, a deletion and a coupon, when the recorded changes are listed, then the coupon and the corrections come newest first by when each was recorded, each with its postings, other first versions are left out, and passing the last one's record time and row ID returns the rest.