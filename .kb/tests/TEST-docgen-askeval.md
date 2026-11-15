---
title: docgen trims a bench report to a result file, loads the committed results and renders the results page
sourceFile: tools/docgen/askeval_test.go
verification_scope: unit
origin:
  kind: agent
  ref: 'AB9 brief: the Ask box evaluation article and the Ask box showing who answered, with the author''s D-222 (2026-11-15)'
  recorded_at: '2026-11-15T09:15:03.978Z'
status: active
id: TEST-docgen-askeval
type: test
---
Table tests of trimReport (a verdict report keeps the host, failing cases and reliability; a model no call answered has no verdict; an old report has none and counts cases under pass^5; a subset is diagnostic), of `docgen askeval import` (the file name sorts in run order, a bad run name is refused), of the committed results (every one complete, newest first) and of the page (the standing table lists only full runs with a verdict, diagnostic and unanswered runs are marked, the hash and schema are shown).