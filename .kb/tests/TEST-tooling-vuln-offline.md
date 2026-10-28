---
title: vulncheck.sh fails on a reported vulnerability and when the registry is unreachable
status: active
sourceFile: scripts/tests/vulncheck.test.sh
origin:
  kind: agent
  recorded_at: '2026-10-27T20:35:54.605Z'
id: TEST-tooling-vuln-offline
type: test
---
Runs scripts/vulncheck.sh in a copy with only the npm lockfile, so govulncheck never runs. With a stub npm whose audit reports a high-severity finding, it exits non-zero and prints the finding. With the registry pointed at a refused localhost port, it exits non-zero and reports that npm audit failed. Neither case needs the network.
