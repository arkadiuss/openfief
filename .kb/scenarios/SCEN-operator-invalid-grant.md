---
title: An invalid PluginGrant is reported and enforces nothing
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T07:00:00.475Z'
id: SCEN-operator-invalid-grant
type: scenario
---
Given a PluginGrant whose egress host is the wildcard *.example.com, when the operator reconciles it and its Plugin, then the grant's Valid condition is False, the Plugin's Ready is False with reason InvalidGrant, and nothing runs.