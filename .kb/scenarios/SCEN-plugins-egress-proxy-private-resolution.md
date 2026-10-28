---
title: A listed host that resolves to a private address is refused
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:10:03.267Z'
id: SCEN-plugins-egress-proxy-private-resolution
type: scenario
---
Given plugin sample with internal.example.com and mixed.example.com on its egress allow-list, when internal.example.com resolves to 10.0.0.7 and mixed.example.com resolves to a public address and 10.0.0.7, then both CONNECTs are refused with 502 and neither address is dialed. Assumes the proxy resolves names itself and plugins have no DNS.