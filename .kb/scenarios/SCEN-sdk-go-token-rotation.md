---
title: The SDK presents the token the kubelet wrote last
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:00:55.786Z'
id: SCEN-sdk-go-token-rotation
type: scenario
---
Given a plugin whose projected token file the kubelet rewrites when it rotates the token, when the plugin calls the core or opens an egress tunnel after the rotation, then the SDK reads the file again and sends the new token, and when the file is missing or empty the call fails before anything is sent.