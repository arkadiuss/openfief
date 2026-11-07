---
title: A series' fee comes from the reviewed table by its sale month
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T08:06:02.186Z'
id: SCEN-pl-early-redemption-fee-table
type: scenario
---
Given the reviewed fee table, when the fee of a series is asked for, then the fee of the latest row of its bond type on or before its sale month is returned, and a series sold before the type's first row or after the reviewed month gets none.