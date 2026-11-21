---
title: A model question over GET gets the ask-from-the-box card and reaches no model; a command answers; the POST asks
status: active
sourceFile: internal/views/ui/askmodel_page_test.go
verification_scope: integration
origin:
  kind: agent
  ref: 'Security review before the alpha.2 cut, known limit #9: a model question must need POST so a same-site GET can''t spend model budget (the coordinator''s follow-up, 2026-11-21)'
  recorded_at: '2026-11-21T19:04:44.478Z'
id: TEST-ui-ask-model-question-needs-post
type: test
---
Through the UI's handler with a fake core and a model turned on: GET /ask by htmx and GET /?q= with a question that is no command show 'Ask this from the box' and make no admission and no model call; a command over GET answers; the same question as the box's htmx POST is answered by the model. The Turnstile, pass and admission tests post as the box does.