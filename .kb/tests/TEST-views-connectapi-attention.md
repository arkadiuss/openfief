---
title: The read API lists attention items and broken transfers, as the OpenAPI document says
status: active
sourceFile: internal/views/connectapi/attention_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T18:30:48.403Z'
id: TEST-views-connectapi-attention
type: test
---
Serves the read API's handler in process and validates each response against the generated OpenAPI document: GET /v1/attention-items lists two synthetic items with their state, detail, owner_resolves and history, and the open count, and lists none without the attention port; GET /v1/transfers lists a confirmed pair whose deposit was corrected as broken, not confirmed, with the deposit's amount.