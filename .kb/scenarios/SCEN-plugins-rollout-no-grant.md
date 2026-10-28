---
title: A plugin without a grant, or without a granted capability, runs nothing
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:36:21.947Z'
id: SCEN-plugins-rollout-no-grant
type: scenario
---
Given the sample plugin and no grant, when the rollout is planned, then nothing runs and everything it requests is pending. Given a grant that approves only the view capability for a connector plugin, then nothing runs either. Given a grant that drops the egress hosts of a running plugin, then the running version keeps running without them.