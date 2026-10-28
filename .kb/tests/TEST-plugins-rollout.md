---
title: Rollout planning runs the current version, keeps the previous one, or runs nothing
status: active
sourceFile: internal/plugins/rollout_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:36:16.896Z'
id: TEST-plugins-rollout
type: test
---
Table tests in package plugins_test call PlanRollout with the synthetic sample plugin, an upgrade at a second fake digest that adds the egress host new.example.net, full, narrowed, view-only, invalid and mismatched grants, and compare the whole Rollout (outcome, effective permissions of the version that runs, pending permissions) with go-cmp. Cases: no grant, a fully granted current version, an upgrade with and without a previous version, an undeployable current version with and without a previous one, an invalid current version, a narrowed grant applied to the running version, a grant without the requested capability, and the error cases. A false pass would be an upgrade that runs with partial permissions; the expected Run is computed from the previous version.