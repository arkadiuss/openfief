---
title: NBP's archive gives every reference rate change with the day it took effect
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T03:56:25.586Z'
id: SCEN-valuation-nbp-reference-rates
type: scenario
---
Given NBP's published archive of base interest rates, when it is read, then every change of the reference rate comes back once, ordered by the day it took effect, as an exact fraction (5,75 as 0.0575), changes of other rates only are left out, and a document that is not the archive, a bad date or rate, or two changes on one day fail.