---
title: The kibi-core patch removes duplicate triples and reports how many
status: active
sourceFile: scripts/tests/kibi-dedupe.test.sh
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-07T19:27:54.016Z'
id: TEST-tooling-kibi-dedupe
type: test
---
Loads the installed, patched kibi-core with kibi's own SWI-Prolog and calls kb_dedupe_graph_triples/1 on a small in-memory graph: with one entity's triples and one relationship stored twice it keeps one copy of each triple, reports 3 triples and 1 entity in kb_dedupe_report/2 and prints the counts on stderr; a clean graph is left alone with no report and no line. It fails when the installed kibi-core lacks the patch.