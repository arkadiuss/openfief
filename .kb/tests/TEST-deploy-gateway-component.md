---
title: The Gateway component routes the owner's hostname to the core's browser port, and the bases expose nothing
status: active
sourceFile: scripts/tests/gateway.test.sh
origin:
  kind: agent
  recorded_at: '2026-10-29T20:58:08.408Z'
id: TEST-deploy-gateway-component
type: test
---
Copies deploy/ to a scratch directory, extracts the overlay docs/guides/expose.md shows (OpenFief's bases plus deploy/gateway, with the owner's Gateway, hostname and Envoy Gateway proxies) and builds it with the pinned kustomize: exactly one HTTPRoute, on the owner's Gateway and hostname, sends every path to Service openfief-core port 8080, which is the Service's http port; the added NetworkPolicy admits the proxies to port 8080 alone; the core's --auth-origin list holds localhost and the route's origin. The core base alone has no HTTPRoute, Ingress or proxy policy, admits plugin pods to port 8081 only, and its connect port targets 8081.