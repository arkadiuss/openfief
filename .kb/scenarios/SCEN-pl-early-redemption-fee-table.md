---
title: A series' fee comes from the reviewed table by sale month
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T08:06:02.186Z'
id: SCEN-pl-early-redemption-fee-table
type: scenario
---
Given the reviewed fee table, when the fee of a series is asked for, then the fee of the latest row of its bond type on or before its sale month is returned; a series sold after the reviewed month gets that type's latest fee, marked assumed, and a series sold before the type's first row gets none.