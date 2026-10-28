---
title: The demo's route through k3s's Traefik reaches the ready core
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T18:00:57.973Z'
id: SCEN-e2e-traefik-route
type: scenario
---
Given a k3s cluster at the demo machine's k3s version with its bundled Traefik, the demo's Ingress for the core without TLS and the demo's core-from-traefik policy, when the client pod sends GET /readyz to Traefik with the route's host, then Traefik answers 200 from the core within three minutes rather than 503 no available server.