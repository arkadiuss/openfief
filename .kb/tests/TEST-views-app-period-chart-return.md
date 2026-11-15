---
title: A period's card return and the performance chart opened on its opening close are the same figure, for 2025 and for November 2025
status: active
origin:
  kind: agent
  ref: 'Design session, on the author''s behalf: D-223, a period''s return runs from the close of the day before its first day and the chart opens there, 2026-11-15'
  recorded_at: '2026-11-15T18:11:56.978Z'
sourceFile: internal/views/app/periodchart_test.go
verification_scope: unit
id: TEST-views-app-period-chart-return
type: test
---
Views.PeriodChange and Views.Performance over in-memory ports with an account valued every day from 2024-12-20 through 2025-12-31 by a fixed pattern, with a second deposit inside the year. For 2025 and for November 2025 the card's opening is the day before the first day; the chart opened on that day from the preset's date and from the card link's date starts at index 100 on it and ends on an index whose distance from 100 equals the card's return to the two places the card shows, compared as decimals, never floats; a chart opened on the period's own first day returns something else, so the test tells the two openings apart. A false pass would need the fixture's valuations to make the two openings agree, which the last check rules out.