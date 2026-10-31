---
title: The read API lists attention items as the OpenAPI document says
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T18:31:23.228Z'
id: SCEN-views-attention-items
type: scenario
---
Given an open import item and a cleared setting item, GET /v1/attention-items lists both with their state, detail, owner_resolves and history and counts one open, and the response matches the generated OpenAPI document.