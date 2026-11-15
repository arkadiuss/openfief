---
title: docgen renders the guide's results table, orders it, finds comments and fails the check when it is stale
status: active
origin:
  kind: agent
  ref: 'The author''s request on the models docs: a table with results, not prose only, eventually comments (2026-11-15)'
  recorded_at: '2026-11-15T19:31:37.797Z'
sourceFile: tools/docgen/askevaltable_test.go
verification_scope: unit
id: TEST-docgen-askeval-table
type: test
---
Table tests of the guide's results table: rows sort passes first, then cases passed, then cost with a missing cost last; only the latest full non-void run of each model and host stands, and void, unanswered and diagnostic runs are left out; every column renders, with the safety count from the failing cases; a comment is found by model@host, then model@provider, then the model alone, and a missing one is empty; consecutive passes stop at a failure or another configuration; the fences replace only the text between them; a void run reads the same on both pages; the safety total equals the bench's cases; the committed guide's table is current; and a stale table fails the check.