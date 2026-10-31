---
title: check-docs.sh rejects undefined tags, edited ground rules and duplicate IDs
status: active
sourceFile: scripts/tests/check-docs.test.sh
origin:
  kind: agent
  recorded_at: '2026-10-27T20:35:13.550Z'
id: TEST-tooling-docs-check
type: test
verification_scope: integration
---
Runs scripts/check-docs.sh against a throwaway git copy of the design documents: unchanged docs pass; an undefined OQ-999 reference, an edited GR-14 line and a second definition of GR-1 each fail and are named.
