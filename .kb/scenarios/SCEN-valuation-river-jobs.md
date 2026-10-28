---
title: One River job per account runs the coalesced revaluation
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T07:56:26.706Z'
id: SCEN-valuation-river-jobs
type: scenario
---
Given three enqueues of acc-1 from 2026-03-10, 2026-03-01 and 2026-03-05, then there is one waiting job and the pending date is 2026-03-01; when the worker runs, then it revalues from that date, clears the pending date and completes the job in one transaction, leaving nothing pending; when a missing input arrives later, then a new job revalues from its date.