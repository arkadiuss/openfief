---
title: A reported vulnerability fails the check
status: active
origin:
  kind: agent
  recorded_at: '2026-10-27T20:32:37.972Z'
id: SCEN-tooling-vuln-finding
type: scenario
---
Given npm audit reports a high-severity vulnerability, when task vuln runs, then it prints the findings and exits non-zero.
