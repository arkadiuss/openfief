---
title: Attention items change by their state rules
status: active
sourceFile: internal/attention/rules_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T18:30:31.798Z'
id: TEST-attention-rules
type: test
---
Pure table tests of the attention domain: Observe raises, changes, reopens and clears an item for each state and finding, keeping an owner-resolved item resolved while the detail is the same; Reconcile keeps one item per subject, raises new conditions, clears vanished ones in subject order and refuses a repeated subject or another kind; Acknowledge and Resolve allow only the owner's transitions and only the kinds the owner may resolve; conditions validate their kind, subject and detail. A false pass would need the expected changes in the tables to be wrong.