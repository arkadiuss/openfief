---
title: A sync tells the connector each account's payout setting
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T07:44:21.493Z'
id: SCEN-ingest-sync-carries-payout
type: scenario
---
Given an account paid out since a day bound to an instance, when the core syncs the instance, then the request carries the account's payout setting and its day, which the plugin reads back intact; an unknown setting or a paid-out one without a valid day is refused as malformed.