---
title: The tools' period arguments map to the typed commands' queries
status: active
origin:
  kind: agent
  ref: 'Design session decision, on behalf of the author: the Ask box performance and value tools and commands answer for a bounded period, a gap against D-194 and D-203, 2026-11-15'
  recorded_at: '2026-11-15T13:36:12.713Z'
id: SCEN-ui-ask-tools-take-a-period
type: scenario
---
Given the performance tool called with a month or with from and to, and the value tool with a period, then each runs the query of the typed command of the same period; a from without a to, a bad day and a value with both a date and a period are refused with a fixed text that quotes nothing.