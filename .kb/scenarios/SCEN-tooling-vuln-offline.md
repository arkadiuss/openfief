---
title: Unreachable vulnerability database fails the check
status: active
origin:
  kind: agent
  recorded_at: '2026-10-27T20:32:52.769Z'
id: SCEN-tooling-vuln-offline
type: scenario
---
Given the npm registry is unreachable, when task vuln runs, then it reports that npm audit failed and exits non-zero, locally and in CI alike.
