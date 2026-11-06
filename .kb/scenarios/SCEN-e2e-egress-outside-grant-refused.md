---
title: The egress proxy refuses a granted plugin request for a host outside its grant
status: active
tags:
  - acceptance:p10-10
  - acceptance:p10-16
origin:
  kind: agent
  recorded_at: '2026-11-06T20:29:59.315Z'
id: SCEN-e2e-egress-outside-grant-refused
type: scenario
---
Given the nbp-fx connector granted api.nbp.pl, which its manifest requests, and example.org, which it does not, when from inside its pod a CONNECT is sent through the egress proxy with a token of its ServiceAccount for the proxy to example.org, to a TEST-NET address and to the core in-cluster service, and a CONNECT is sent without a token and with the token of a plugin granted no host, then the proxy refuses each with 407 and opens no tunnel; a plain HTTP request through the proxy is refused with 407 too. No request goes to a real external service. Runs on both kind and k3d.