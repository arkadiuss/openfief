---
title: The owner changes a bond account's payout setting from a day
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T16:48:02.467Z'
id: SCEN-ingest-bond-payout-change-from-day
type: scenario
---
Given a hand-entered bond account, when the owner creates it with a payout setting, then the setting covers its whole history; when the owner changes it from a day given, today when none is, then a later day is refused as invalid and the payout withdrawals booked from that day on that the new setting no longer makes are superseded by deletions, while earlier ones stay.