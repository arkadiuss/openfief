---
title: vulncheck.sh warns locally and fails in CI when the registry is unreachable
status: active
sourceFile: scripts/tests/vulncheck.test.sh
origin:
  kind: agent
  recorded_at: '2026-10-27T20:35:54.605Z'
id: TEST-tooling-vuln-offline
type: test
---
Runs scripts/vulncheck.sh in a copy with only the npm lockfile and the registry pointed at a refused localhost port: without CI it warns and exits zero, with CI=true it exits non-zero and says why.
