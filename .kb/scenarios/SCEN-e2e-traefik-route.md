---
title: The demo's route through k3s's Traefik reaches the read API on the core
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T18:00:57.973Z'
id: SCEN-e2e-traefik-route
type: scenario
---
Given the demo's Ingress on k3s's Traefik and the core-from-traefik policy, when a client in another namespace asks /v1/accounts through Traefik without a token, then the core itself answers 401, not Traefik's 503 or the UI's 404, once Traefik has picked up the route.