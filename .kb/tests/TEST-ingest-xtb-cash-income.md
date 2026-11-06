---
title: XTB's free-funds interest and closed CFD results map to cash income
status: active
sourceFile: internal/ingest/xtb/cashincome_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-02T05:50:47.355Z'
id: TEST-ingest-xtb-cash-income
type: test
---
Table tests over synthetic xStation exports: a Free funds interest row is booked as cash income with the Free funds interest tax rows of its date summed as the tax withheld, or none; a tax row with no interest row of its date, matching two, or positive is an ingest problem and the interest is booked without it; negative interest is unsupported. A closed CFD position's Profit/Loss is cash income on its close date, keyed closed-position plus its position ID, a loss taking cash out; one of no result is skipped and one without a position ID unsupported. A false pass would need the mapper to book interest as a deposit or drop the tax (D-152, D-24).