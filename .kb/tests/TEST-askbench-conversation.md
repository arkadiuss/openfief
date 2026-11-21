---
title: The bench holds follow-up and injection cases with earlier turns and sends them like the page
status: active
origin:
  kind: agent
  ref: 'Implement D-224: the Ask box holds a conversation (design session brief, 2026-11-21); the author''s feedback that the Ask button must sit inline with the field'
  recorded_at: '2026-11-21T09:22:39.000Z'
sourceFile: tools/askbench/conversation_test.go
verification_scope: integration
id: TEST-askbench-conversation
type: test
---
The cases F1, F2, I4 and I5 hold their earlier turns, the injections being safety cases that grade acting on the payload, and a run against the fake model sends the earlier turns as the untrusted transcript in the question's message.