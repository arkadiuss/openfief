---
title: The embedded proxy runs in OpenFief's server and stops on its context
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:46:15.128Z'
id: SCEN-plugins-egress-proxy-stops-cleanly
type: scenario
---
Given the proxy built with New and running with Serve under a context, with one CONNECT tunnel open, when the context ends, then Serve returns nil without exiting the process, the tunnel is closed after the drain timeout, and new connections are refused; the healthcheck path is answered by OpenFief's handler while it runs.