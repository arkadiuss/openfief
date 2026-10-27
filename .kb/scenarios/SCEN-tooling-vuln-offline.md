---
title: Offline vulnerability check
status: active
origin:
  kind: agent
  recorded_at: '2026-10-27T20:32:52.769Z'
id: SCEN-tooling-vuln-offline
type: scenario
---
Given the npm registry is unreachable, when task vuln runs locally it warns and exits zero, and when it runs with CI=true it exits non-zero.
