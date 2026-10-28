---
title: Egress requests carry the current token in the proxy URL
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:00:52.566Z'
id: SCEN-sdk-go-egress-proxy-url
type: scenario
---
Given HTTPS_PROXY names the egress proxy and OPENFIEF_EGRESS_TOKEN_FILE names the egress token file, when a plugin sends an https request with the SDK's egress client, then it opens a CONNECT tunnel whose proxy URL has the token current at that moment as password, sent as Proxy-Authorization; after the kubelet rotates the token the next request carries the new one; a plain http request, a missing proxy or a missing token file is refused without contacting anyone.