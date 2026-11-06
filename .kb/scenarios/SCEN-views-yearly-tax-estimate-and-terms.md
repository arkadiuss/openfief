---
title: The read API says when a yearly tax is an estimate and what each wrapper asks of an account
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T20:11:32.602Z'
id: SCEN-views-yearly-tax-estimate-and-terms
type: scenario
---
Given OKI's yearly tax counted through a day before the year's end, when the read API serves it, then it carries estimate with the whole-złoty base and tax, and the built-in UI marks it Estimate; the settings list each wrapper's account forms and whether it takes earlier payments, and each account its form.