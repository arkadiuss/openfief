---
title: The UI presents its plugin token to the AuthService and the session to the read API
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:48:54.800Z'
id: SCEN-ui-core-credentials
type: scenario
---
Given the plugin's token file, when the UI calls the AuthService, then each call carries the token read from the file at that moment, also after it rotates, or, outside a cluster, the OPENFIEF_AUTH_CALLER_TOKEN token when that is set; when it calls the read API, each call carries the owner's session token as its bearer and never the plugin's token.