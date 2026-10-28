---
title: 'docgen fresh: glob matching, git log parsing, the staleness gate and package coverage'
status: active
sourceFile: tools/docgen/fresh_test.go
verification_scope: unit
verification_perspective: internal
origin:
  kind: agent
  recorded_at: '2026-10-28T09:32:38.295Z'
id: TEST-docs-fresh
type: test
---
Table tests in tools/docgen/fresh_test.go: path globs with * and **; git log records with and without a Docs-Unchanged trailer; covered changes with and without their page, later in the push, exempt by trailer, tests and test data ignored; packages no page covers; front matter parsing; and the staged mode end to end over an in-memory repository.
