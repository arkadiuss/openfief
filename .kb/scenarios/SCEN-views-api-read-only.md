---
title: Nothing can be written through the read API
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:43:13.820Z'
id: SCEN-views-api-read-only
type: scenario
---
Given the read API, when a client sends POST, PUT, PATCH or DELETE to any REST path under /v1, with or without a valid token, then the answer is 405 Method Not Allowed with Allow: GET and nothing is read or changed; and the service itself defines only unary methods that declare no side effects and map to GET without a body, and the generated OpenAPI document has only GET operations, so no handler can accept data.