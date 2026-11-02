---
title: A closed CFD's result is cash income; open CFDs stay flagged
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T05:51:22.413Z'
id: SCEN-ingest-xtb-cfd-closed-cash-income
type: scenario
---
Given an xStation export with a closed CFD position and an open one, when it is mapped, then the closed position's Profit/Loss is cash income on the UTC date it closed, keyed closed-position plus its position ID, a loss taking cash out, while the open CFD position and a Cash Operations row of a CFD category are unsupported ingest problems.