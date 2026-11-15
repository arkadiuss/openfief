---
title: The overview's time-weighted return link opens the performance chart on the year's opening close
status: active
origin:
  kind: agent
  ref: 'Design session, on the author''s behalf: D-223, a period''s return runs from the close of the day before its first day and the chart opens there, 2026-11-15'
  recorded_at: '2026-11-15T18:11:59.737Z'
sourceFile: internal/views/ui/home_test.go
verification_scope: integration
id: TEST-ui-period-chart-link
type: test
---
The overview over a fake core behind real Connect servers: the value card's time-weighted return of this year links to the Performance tab with the start parameter set to the opening day of the year's return, the last valued day on or before 31 December, here the first day of the fake series because it starts later. A false pass would need the fake's series to be shaped to the link, but it is the shared fake the other overview tests read.