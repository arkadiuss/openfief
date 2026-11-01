---
title: A sync carries every payout setting with its day
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T16:48:06.531Z'
id: SCEN-ingest-sync-carries-payout-history
type: scenario
---
Given a bound bond account whose payout setting changed, when the core syncs its connector, then the request carries the current setting with its day and the earlier ones, the plugin refuses settings out of order, and the connector pays out each coupon and maturity by the setting in force on its day.