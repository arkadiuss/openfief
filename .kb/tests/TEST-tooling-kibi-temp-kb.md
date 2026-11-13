---
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T02:54:48.005Z'
title: The staged check keeps its throwaway store off the disk
sourceFile: scripts/tests/kibi-temp-kb.test.sh
verification_scope: integration
id: TEST-tooling-kibi-temp-kb
type: test
---
Runs kibi's own createTempKb and projectStagedEntities on 300 synthetic requirements, so the projection spans three batches: the throwaway store's kb.rdf must stay empty (kibi's wrapper saved the whole store after each batch, which made kibi check --staged quadratic in the knowledge base), the 300 requirements must be queryable in the process, and cleanupTempKb must remove the directory. Without the kibi-cli patch the test fails with 205354 bytes written.