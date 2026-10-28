---
title: Offline vulnerability check
status: active
origin:
  kind: agent
  recorded_at: '2026-10-27T20:32:52.769Z'
id: SCEN-tooling-vuln-offline
type: scenario
---
Given the npm registry refuses connections, or an egress proxy refuses vuln.go.dev with 403 Forbidden, when task vuln runs locally it warns that the database is unreachable and exits zero, and when it runs with CI=true it exits non-zero. Assumes govulncheck reports every database download failure as 'fetching vulnerabilities' and exits 3 only for findings.