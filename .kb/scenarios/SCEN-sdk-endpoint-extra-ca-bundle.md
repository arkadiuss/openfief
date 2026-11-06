---
title: An extra CA bundle is trusted besides the system roots, never instead
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T21:55:01.850Z'
id: SCEN-sdk-endpoint-extra-ca-bundle
type: scenario
---
Given two CAs made for the test, one standing in for the system roots, when the endpoint has the other CA as its extra bundle then servers signed by either are trusted through the egress proxy; without a bundle only the system roots apply; a server signed by neither is refused as an unknown authority; a missing bundle Secret adds nothing and a bundle without a valid PEM certificate is refused.