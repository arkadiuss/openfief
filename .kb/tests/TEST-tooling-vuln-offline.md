---
title: vulncheck.sh fails on a reported vulnerability and when a vulnerability database is unreachable
status: active
sourceFile: scripts/tests/vulncheck.test.sh
origin:
  kind: agent
  recorded_at: '2026-10-27T20:35:54.605Z'
id: TEST-tooling-vuln-offline
type: test
---
Runs scripts/vulncheck.sh in a throwaway copy with the npm lockfile, so nothing touches the network. Case one: a stub npm whose audit reports a high-severity finding; the check exits non-zero and prints the finding. Case two: the registry on 127.0.0.1:9, which refuses connections; the check exits non-zero and reports that npm audit failed. Case three: a go.mod, a stub govulncheck that fails the way it does when a proxy refuses vuln.go.dev (fetching vulnerabilities ... Forbidden) and a clean npm audit; the check exits non-zero and reports that govulncheck failed. There is no CI switch: the same result is expected locally and in CI.
