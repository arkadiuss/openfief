---
title: The Ask box with a model answers through the page with the asker's session, the cards it read and what was withheld, and behaves as before without one
status: active
sourceFile: internal/views/ui/askmodel_page_test.go
verification_scope: integration
origin:
  kind: agent
  ref: M2 brief (the Ask box's model, D-203) on the model router, 2026-11-13
  recorded_at: '2026-11-13T18:11:34.621Z'
id: TEST-ui-ask-model-page
type: test
---
Serves the UI over fake Connect services, one standing for the core's router: the page for a model's answer, a command answered at once, a withheld card, each fallback, the cards for a model that can't answer, and the log.