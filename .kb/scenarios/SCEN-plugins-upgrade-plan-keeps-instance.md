---
title: An upgrade keeps the instance and grants exactly what the new version asks
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T17:03:59.427Z'
id: SCEN-plugins-upgrade-plan-keeps-instance
type: scenario
---
Given an installed instance, when it is upgraded to a new signed image, then the Plugin keeps its name, its config unless the owner gives a new one, and its in-cluster service while the new manifest still asks for one; the grant approves exactly what the new version requests, and the summary lists what it adds and no longer asks for.