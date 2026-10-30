---
title: The install lets the operator read the one picked Service, and uninstall takes that back
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T15:10:21.871Z'
id: SCEN-plugins-cli-in-cluster-service-access
type: scenario
---
Given an install with an in-cluster service, when the owner confirms, then the CLI checks the Service exists with that port and a selector and creates in its namespace a Role allowing only get on that Service and a RoleBinding to the operator's ServiceAccount; uninstalling the instance deletes both.