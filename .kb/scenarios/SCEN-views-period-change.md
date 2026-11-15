---
title: A period of days gets its start and end values, change, return and deposits from the core, or a refusal
status: active
origin:
  kind: agent
  ref: 'Design session, on the author''s behalf: the Ask box''s performance and value tools answer for a bounded period, a gap against D-194 (D-203: the read API answers first), 2026-11-15'
  recorded_at: '2026-11-15T12:27:08.778Z'
id: SCEN-views-period-change
type: scenario
---
Given a portfolio whose account was funded on the first day and holds a deposit in the middle of a period, when the period's change is asked for from a day to a later day, then the gross, net and liquidation values at the end of the day before and at the end of the last day come with their change in money and in percent, the time-weighted return leaves the deposit out and the deposit is listed as paid in. A period that starts before the record starts on its first day, a value missing at either end is left unset, and a period ending after the day every account is valued through, one before the record, and a reversed one are refused. Assumes the performance fixture's synthetic accounts and rates.