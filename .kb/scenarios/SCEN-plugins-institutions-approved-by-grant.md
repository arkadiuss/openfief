---
title: A connector books only for institutions its grant approves
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T07:22:20.251Z'
id: SCEN-plugins-institutions-approved-by-grant
type: scenario
---
Given a connector whose signed manifest declares institutions, when its PluginGrant doesn't approve one of them, then the version waits for approval with that institution in status.pendingApproval; when the grant approves them, it runs with exactly those both name in status.effective.institutions, and the core refuses a delivery naming any other institution, as it refuses an undeclared one (D-182).