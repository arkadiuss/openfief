---
title: A ROR or DOR period takes the reference rate in force ten business days before its month
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T03:45:50.385Z'
id: SCEN-pl-reference-rate-day
type: scenario
---
Given NBP's reference rate changes and a month in which ROR or DOR interest periods start, when the reading for that month is picked, then it is the rate of the latest change in force on the tenth business day before the month's first day, where business days skip Saturdays, Sundays and the holidays the act on days off work lists for that year (6 January from 2011, 24 December from 2025, Easter, Corpus Christi); dates before 1991 are refused.