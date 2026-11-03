---
title: A payout setting keeps the settings in force before a change
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T11:49:34.044Z'
id: SCEN-ingest-payout-setting-history
type: scenario
---
Given a bond account paid out since 2026-03-02 with cash before, when it changes back to cash from 2026-03-09, then a coupon due on 2026-03-05 is paid out, one due on 2026-03-10 stays as cash, and one due before 2026-03-02 stays as cash; the setting reads back unchanged from the ingest schema.