---
title: The Ask box evaluation guide's results table is generated and checked
status: active
origin:
  kind: agent
  ref: 'The author''s request on the models docs: a table with results, not prose only, eventually comments (2026-11-15)'
  recorded_at: '2026-11-15T19:31:39.000Z'
id: SCEN-askeval-guide-table
type: scenario
---
Given the committed results, the bench's history and the comments file, when docgen runs askeval table, then the guide holds one row per model and host with its latest full run that is not void, ordered passes first, then cases passed, then cost, each with its verdict and scorer version, cases passed, safety cases passed, transport failure rate, cost, consecutive passes, last run and comment, a model with no comment has an empty one, a void run and a diagnostic run have no row, and with -check the command fails while the guide's table differs.