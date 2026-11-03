---
title: Each Secret role becomes the instance's own Secret, joined with two hyphens
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T18:23:43.982Z'
id: SCEN-plugins-requested-secret-names-double-hyphen
type: scenario
---
Given a plugin instance whose manifest names Secret roles, when its requested permissions are computed and its workload compiled, then it requests and mounts <instance>--<role> for each role, and a role that is no DNS label or that holds two hyphens in a row is refused when the manifest is validated.