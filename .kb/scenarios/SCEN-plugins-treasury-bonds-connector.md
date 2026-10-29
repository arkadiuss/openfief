---
title: The treasury-bonds plugin maps lots for the core only and needs no host or secret
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T01:14:02.406Z'
id: SCEN-plugins-treasury-bonds-connector
type: scenario
---
Given the plugin serving the connector contract behind the SDK's core-token check, when the core calls ImportBondLots, then it maps the lots into its configured account as of today in Warsaw, or answers INVALID_ARGUMENT naming the rule; a call without the core's token is refused; Sync delivers nothing; its manifest asks only for the connector capability.