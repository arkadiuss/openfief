---
title: Invalid inputs are rejected and nothing is stored
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T07:29:07.114Z'
id: SCEN-valuation-input-validation
type: scenario
---
Given a price with no instrument or a zero close, an FX rate for PLN or below zero, an index reading of an unknown kind or for month 13, bond terms without an ISIN or series, or an unknown tax wrapper, when it is recorded, then recording fails with an invalid-input error and nothing is stored or enqueued.