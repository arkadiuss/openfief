---
title: A Plugin without a PluginGrant gets no workload
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:59:52.046Z'
id: SCEN-operator-no-grant
type: scenario
---
Given a Plugin named nogrant requesting api.example.com and no PluginGrant named nogrant, when the operator reconciles it, then no Deployment or ServiceAccount exists, its NetworkPolicy denies everything, Ready is False with reason NoGrant, and status.pendingApproval lists its capability, host and Secret. Given the grant is deleted later, then the running Deployment, ServiceAccount and allow-list entry are removed.