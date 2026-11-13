---
title: After the core is ready, a SpiceDB outage fails authorization closed with 503 and does not make the core not ready; it is logged and counted once each way
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T15:35:05.391Z'
id: SCEN-core-authz-outage-fails-closed-ready-at-start
type: scenario
---
Given a core on --authz=spicedb that has the key and whose start's reconcile has succeeded, so SpiceDB has answered and holds the schema, when SpiceDB stops, then /readyz still answers 200 and /healthz too, every request that needs an authorization check answers 503 with the text authorization is unavailable, never 500, 403, 404 or an answer, and the core logs once that authorization is unavailable and counts it in a metric; when SpiceDB is back, then the requests answer again, the core logs once that it is available again and counts it; and before the key is there, or before SpiceDB has answered once with its schema written, /readyz answers 503 saying which it waits for. Assumes the Service's readiness probe is /readyz alone.