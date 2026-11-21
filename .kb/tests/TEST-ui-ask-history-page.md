---
title: 'Through the page: a POSTed follow-up reaches the model with its history, the server caps what it is sent, nothing is logged and the model has the MCP toolbox'
status: active
origin:
  kind: agent
  ref: 'Implement D-224: the Ask box holds a conversation (design session brief, 2026-11-21); the author''s feedback that the Ask button must sit inline with the field'
  recorded_at: '2026-11-21T09:22:39.000Z'
sourceFile: internal/views/ui/askhistory_page_test.go
verification_scope: integration
id: TEST-ui-ask-history-page
type: test
---
Through the UI over its fake core: a POST to /ask with a history reaches the model as a transcript and the answer carries its own turn as data-turn; a command inside a conversation is answered by the box; twelve long turns reach the model as at most six and 4,000 characters; a history that is no array is dropped; a body over 64 KiB is refused with 413; a viewer asks by POST and a cross-site POST is refused; a unique marker in the question, the answer and the earlier turns is in no log line even when the model fails; the tools the model is offered equal the shared toolbox the MCP server serves, by name, description and schema; and a follow-up with no Turnstile token is told a check is needed and, with a good token in its body, goes on with its history.