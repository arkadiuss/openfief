---
title: A connector's reported broker accounts are recorded before anything is booked
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T11:35:53.411Z'
id: SCEN-ingest-broker-account-record
type: scenario
---
Given a delivery of instance ibkr-main from ibkr that lists the synthetic broker accounts U0000001 and U0000002, when the core receives it, then both are recorded for the instance with the institution, today as last seen and as in the last delivery, unmapped when new; a later delivery that lists only U0000001 keeps U0000002 but no longer in the last delivery.