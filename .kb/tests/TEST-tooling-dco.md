---
title: dco-check.sh passes signed-off commits and names unsigned ones
status: active
sourceFile: scripts/tests/dco-check.test.sh
origin:
  kind: agent
  recorded_at: '2026-10-28T04:51:20.011Z'
id: TEST-tooling-dco
type: test
---
Builds a throwaway repository and runs scripts/dco-check.sh on it: a commit signed off by its author passes; a commit signed off by someone else and a commit with no sign-off each fail with the commit named in the output.
