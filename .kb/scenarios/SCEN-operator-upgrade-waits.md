---
title: An upgrade requesting a new host waits, then rolls out once approved
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:59:54.204Z'
id: SCEN-operator-upgrade-waits
type: scenario
---
Given a granted Plugin running the first fake digest, when its spec moves to a second digest that also requests new.example.net:8443, then the Deployment keeps the first digest, ApprovalRequired is True and status.pendingApproval lists the new host. When the PluginGrant adds that host, then the Deployment runs the second digest and the allow-list holds both hosts.