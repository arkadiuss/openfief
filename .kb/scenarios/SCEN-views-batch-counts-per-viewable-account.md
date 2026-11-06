---
title: A viewer of one account sees a shared batch's counts of that account alone
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T11:51:16.652Z'
id: SCEN-views-batch-counts-per-viewable-account
type: scenario
---
Given an upload and a sync booked into acc-1 and acc-2 with counts per account, when a viewer of acc-1 lists the activity and the sync's attempts, then each names acc-1 alone with its counts, the upload's totals are acc-1's and it shows no left-out rows; the owner sees both accounts and the totals.