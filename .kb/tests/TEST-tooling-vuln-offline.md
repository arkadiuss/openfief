---
title: vulncheck.sh warns locally and fails in CI when a vulnerability database is unreachable
status: active
sourceFile: scripts/tests/vulncheck.test.sh
origin:
  kind: agent
  recorded_at: '2026-10-27T20:35:54.605Z'
id: TEST-tooling-vuln-offline
type: test
---
Runs scripts/vulncheck.sh in a throwaway copy of the repository, so nothing touches the network. Case one: only the npm lockfile and a registry on 127.0.0.1:9, which refuses connections. Case two: a go.mod plus a stub govulncheck on PATH that fails the way govulncheck does when an egress proxy refuses vuln.go.dev (fetching vulnerabilities ... Forbidden). Each case runs without CI, where the script must warn and exit zero, and with CI=true, where it must exit non-zero and say the database was unreachable. A false pass would be a script that treats any non-zero govulncheck exit as unreachable: findings exit 3 and are excluded before the match.