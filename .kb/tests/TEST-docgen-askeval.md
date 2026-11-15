---
title: docgen trims a bench report to a result file, loads the committed results and renders the results page
sourceFile: tools/docgen/askeval_test.go
verification_scope: unit
origin:
  kind: agent
  ref: 'Design session decisions (author''s delegate): a scorer version and offline re-scoring of the Ask box bench, with D-217, D-219, D-220 and D-221, 2026-11-15'
  recorded_at: '2026-11-15T15:39:11.487Z'
status: active
id: TEST-docgen-askeval
type: test
---
Table tests of trimReport (a verdict report keeps the host, failing cases, reliability and scorer version; a model no call answered has no verdict; an old report has none and counts cases under pass^5 and has no scorer version; a subset is diagnostic), of `docgen askeval import` (the file name sorts in run order, a bad run name is refused), of the committed results (every one complete, newest first) and of the page (the standing table lists only full runs with a verdict, diagnostic and unanswered runs are marked, the hash and schema are shown, and every verdict carries the scorer version, a file without one reading as scorer 1 before versioning).