---
title: A sync reconciles activities and cash balances and settles
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T01:20:58.522Z'
id: SCEN-views-ghostfolio-reconcile-cash
type: scenario
---
Given owned accounts whose activities and balances differ from the projection in every way (missing, edited, copied, moved, stray, on a leftover asset of an earlier version, balances missing, with another value or on another date), when the plan is applied the way Ghostfolio behaves, then a second plan is empty, the owned accounts hold exactly the desired activities and balances, no import lands on a leftover asset of the same name, and the owner's own activities and balances are untouched.