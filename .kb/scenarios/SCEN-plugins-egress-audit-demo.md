---
title: After each demo deploy, a Job audits the egress proxy's log against the PluginGrants
status: active
tags:
  - acceptance:p10-16
origin:
  kind: agent
  recorded_at: '2026-11-07T03:30:48.504Z'
id: SCEN-plugins-egress-audit-demo
type: scenario
---
Given a demo deploy, when task deploy:demo renders the signed artifact, then its audit path holds a Job that runs egress-proxy audit with the egress proxy's verified digest and the deploy's tag, as a ServiceAccount whose Role only lists pods and grants and reads pod logs, which Flux re-creates after the app layer and waits for, so the demo's real traffic to NBP is checked against its grant and a failed audit fails that layer.