---
title: A connector accepts calls only from the core, with the core's token for that instance
status: active
tags:
  - acceptance:p10-17
origin:
  kind: agent
  recorded_at: '2026-11-06T20:30:01.412Z'
id: SCEN-e2e-connector-refuses-bad-core-token
type: scenario
---
Given the XTB connector running, when the core pod calls it with a token of the core ServiceAccount whose audience is that instance, then the call passes the connector auth check and is answered as unimplemented; when the core pod calls it without a token, with a malformed token, with the core token for another instance or with the plugin own token for that audience, then each call is refused with 401; and when the client pod, the outsider pod or another plugin pod calls it, even with the valid token, no connection is made. Runs on both kind and k3d.