---
title: check-docs.sh rejects undefined tags and edited ground rules
status: active
sourceFile: scripts/tests/check-docs.test.sh
origin:
  kind: agent
  recorded_at: '2026-10-27T20:35:13.550Z'
id: TEST-tooling-docs-check
type: test
---
Runs scripts/check-docs.sh against a throwaway git copy of the docs: unchanged docs pass, an undefined OQ-999 reference fails and is named, and an edited GR-14 line fails and is named.
