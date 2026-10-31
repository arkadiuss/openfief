---
title: A broker account is remapped only while nothing was booked from it
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T14:35:11.560Z'
id: SCEN-ingest-remap-unused-broker-account
type: scenario
---
Given a broker account mapped to an account that holds transactions from its connector instance, when the owner remaps it, then it is refused; given one whose account holds none from it, then it maps to the new account, or none, and the instance is bound to the new account.