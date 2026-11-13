---
title: The proxy warns at start, naming the host, whenever an exemption is set
status: active
origin:
  kind: agent
  recorded_at: '2026-11-12T23:20:24.912Z'
id: SCEN-plugins-egress-exemption-warns-at-start
type: scenario
---
Given the proxy started with an exemption, when its log is read, then it holds a warning that a host is exempt from the refused address ranges, with the host and its addresses; and the proxy started without an exemption logs no such warning.