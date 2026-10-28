---
title: Unreachable vulnerability database fails the check
status: active
origin:
  kind: agent
  recorded_at: '2026-10-27T20:32:52.769Z'
id: SCEN-tooling-vuln-offline
type: scenario
---
Given the npm registry refuses connections, or an egress proxy refuses vuln.go.dev with 403 Forbidden, when task vuln runs, then it reports which check failed and exits non-zero, locally and in CI alike.
