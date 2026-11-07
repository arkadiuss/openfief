---
title: With the HTTPRoute component, the core and the UI are reachable over HTTPS on the owner's hostname through a Gateway
status: active
tags:
  - acceptance:p10-13
origin:
  kind: agent
  recorded_at: '2026-11-06T19:44:40.795Z'
id: SCEN-e2e-httproute-https-through-gateway
type: scenario
---
Given the bundle install and a Gateway of the cluster implementation, Envoy Gateway on kind and Traefik on k3d, that terminates TLS for a test hostname with a certificate from a CA made for the run, when the route and NetworkPolicy of deploy/gateway are applied with their placeholders replaced as the expose guide does, then a client that dials the Gateway data plane and names the hostname gets /readyz 200, /v1/accounts 401 and the UI login page, each over HTTPS verified against that CA. The component's patch of the core's auth origins follows once the other bundle tests are done, in SCEN-e2e-passkey-on-route-hostname. Runs on both kind and k3d.