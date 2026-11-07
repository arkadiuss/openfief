---
title: The read API serves a bond series' terms with their sources
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T13:26:33.055Z'
id: SCEN-views-bond-series-terms
type: scenario
---
Given a treasury bond whose terms are stored, when a reader asks for the instrument, then the answer carries its series, first-period rate, margin and early-redemption fee, where the rates came from (published or entered) and where the fee came from (table, assumed or entered), and for published rates the attribution the ministry's licence asks for; an instrument that is no bond has none.