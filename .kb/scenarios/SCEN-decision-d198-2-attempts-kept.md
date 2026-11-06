---
title: Every sync attempt is kept with its class, never its text
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T06:15:01.845Z'
id: SCEN-decision-d198-2-attempts-kept
type: scenario
---
Given sync attempts that succeeded and failed with errors whose text names a path or an amount, when the ingest store records and lists them, then every attempt is kept with its error class and no error text is stored or served.