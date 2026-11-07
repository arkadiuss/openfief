---
title: A fee past the reviewed table is carried forward as assumed
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T13:26:25.751Z'
id: SCEN-pl-fee-carried-forward-assumed
type: scenario
---
Given the reviewed fee table, when the fee of a series is asked for, then a series sold within the reviewed months gets the fee of the latest row of its bond type on or before its sale month, from the table; a series sold after the month the table was reviewed through gets that type's latest reviewed fee, marked assumed, with the row and letter it came from; and a series sold before its type's first row gets none.