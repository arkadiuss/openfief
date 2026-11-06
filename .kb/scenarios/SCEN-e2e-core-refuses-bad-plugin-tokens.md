---
title: The core refuses a plugin call without a valid view plugin token on either port
status: active
tags:
  - acceptance:p10-17
origin:
  kind: agent
  recorded_at: '2026-11-06T20:30:02.106Z'
id: SCEN-e2e-core-refuses-bad-plugin-tokens
type: scenario
---
Given the built-in UI, a view plugin, and the XTB connector running, when the XTB pod calls the core plugin port, then the read API and the AuthService answer a token of the UI ServiceAccount for the core, and refuse a call without a token, with a malformed token, with the connector own token for the core and with the UI token for the egress proxy, the read API with 401 and the AuthService with 403; and when the client pod calls the core browser port with a malformed token or the UI token for the core, the read API refuses it with 401, since that port takes only sessions and script tokens. Runs on both kind and k3d.