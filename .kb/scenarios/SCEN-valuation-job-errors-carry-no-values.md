---
title: A failed revaluation job is logged and stored by its kind only
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T15:15:46.294Z'
id: SCEN-valuation-job-errors-carry-no-values
type: scenario
---
Given a revaluation or end-of-day job whose work fails with an error carrying an amount, an ISIN or an account number, when River records the failure, then River's log line and the job's stored errors name the job and the failure by its known errors or its innermost type only, and River's snooze still reaches River as a snooze (GR-12).