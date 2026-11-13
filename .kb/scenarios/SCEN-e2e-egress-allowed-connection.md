---
title: A granted plugin's CONNECT to the exempt fixture host is allowed and logged, and a granted host in the same private range is refused
status: active
tags:
  - acceptance:p10-16
origin:
  kind: agent
  recorded_at: '2026-11-12T23:20:03.674Z'
id: SCEN-e2e-egress-allowed-connection
type: scenario
---
Given the egress proxy started with the test-only exemption for the suite's egress fixture host (D-215), a NetworkPolicy rule that lets the proxy reach only that fixture's pods on their port, a connector instance granted that host and a second instance granted another Service of the same private range, when the first instance makes one CONNECT to the exempt host through the proxy with its token, then the proxy answers 200 and logs an allowed decision line with the instance as role and the host and port as requested_host, and no error; when the second instance makes a CONNECT to its granted host, then the proxy answers 502 and logs that the address is a forbidden destination, although its NetworkPolicy would allow the connection; the proxy's start log holds the exemption warning naming the exempt host; and the audit of the whole log finds an allowed destination and every allowed line on its instance allow-list, so outbound traffic goes only to configured destinations on an allowed connection. Runs on both kind and k3d.