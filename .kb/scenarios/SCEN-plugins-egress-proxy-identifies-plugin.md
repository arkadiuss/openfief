---
title: The egress proxy accepts only a plugin's egress-audience token
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:09:56.462Z'
id: SCEN-plugins-egress-proxy-identifies-plugin
type: scenario
---
Given a CONNECT whose Proxy-Authorization is Basic with the projected token as the password, as Go's and Node's clients send it for a proxy URL with credentials, and the TokenReview API authenticates that token for audience openfief-egress as system:serviceaccount:openfief:plugin-sample, when the proxy identifies it, then the plugin is sample; with no token, Bearer, a token for openfief-core only, an unauthenticated token or a non-plugin ServiceAccount, the CONNECT is refused with 407 and the token is never logged.