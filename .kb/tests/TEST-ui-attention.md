---
title: The built-in UI lists attention items in one place and takes the owner's actions
status: active
sourceFile: internal/views/ui/attention_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-10-31T18:30:51.254Z'
id: TEST-ui-attention
type: test
---
Runs the UI against fake core services on an httptest server: the Attention page lists both items with links to their pages and history, the navigation and the overview count the open ones, the owner acknowledges one and resolves the other; a read session sees no buttons and is refused; the core's refusal shows; a malformed form never reaches the core; a core that can't count leaves other pages working; and a broken transfer is listed apart on the transfers page with a button to reject it.