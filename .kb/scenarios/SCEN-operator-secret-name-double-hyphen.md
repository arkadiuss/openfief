---
title: An instance's Secret joins its name and role with two hyphens
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T18:23:39.034Z'
id: SCEN-operator-secret-name-double-hyphen
type: scenario
---
Given instance ibkr-flex whose manifest names role token and instance ibkr whose manifest names role flex-token, when each requests and its workload mounts its Secrets, then ibkr-flex asks for ibkr-flex--token and ibkr for ibkr--flex-token, and the operator runs both side by side. Assumes both grants approve exactly what was requested.