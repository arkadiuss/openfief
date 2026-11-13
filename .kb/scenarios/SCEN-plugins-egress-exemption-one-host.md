---
title: The proxy's test-only exemption covers one host at exactly its addresses and nothing else
status: active
origin:
  kind: agent
  recorded_at: '2026-11-12T23:20:24.789Z'
id: SCEN-plugins-egress-exemption-one-host
type: scenario
---
Given the proxy with an exemption for fixture.example.com at 10.0.0.7 and plugin sample with several hosts on its allow-list, when fixture.example.com resolves to 10.0.0.7, then the CONNECT is answered 200 and the address is dialed; when another listed host resolves to 10.0.0.7 or to 10.0.0.8, or fixture.example.com resolves to 10.0.0.8 or to 10.0.0.7 and a further address, or the exempt address is the CONNECT's host, then it is refused (502, or 407 for the IP address) and nothing is dialed; and with no exemption the exempt host itself is refused with 502.