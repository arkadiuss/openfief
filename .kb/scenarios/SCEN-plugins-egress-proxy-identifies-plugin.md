---
title: The egress proxy accepts only a plugin's projected token
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:09:56.462Z'
id: SCEN-plugins-egress-proxy-identifies-plugin
type: scenario
---
Given a CONNECT with Proxy-Authorization Bearer and a token the TokenReview API authenticates for audience openfief-core as system:serviceaccount:openfief:plugin-sample, when the proxy identifies it, then the plugin is sample; with no token, a Basic credential, a token for another audience, an unauthenticated token or a non-plugin ServiceAccount, the CONNECT is refused and the token is never logged.