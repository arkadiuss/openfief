---
title: The core serves the model portfolio service to the UI with the owner's passkey session
status: active
sourceFile: cmd/core/model_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T23:10:18.546Z'
id: TEST-core-model-service
type: test
verification_scope: integration
---
Under task test:pg: a fresh core in a fake cluster whose TokenReview accepts a synthetic token of the view plugin ui, a passkey enrolled through core enroll-link and a software authenticator, then the ModelPortfolioService called with that view plugin token: saving, reading back and refusals for a read session, an unknown session, an origin off the allow-list and a caller that is not a view plugin.