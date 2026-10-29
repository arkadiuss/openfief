---
title: The core serves the model portfolio service to the UI with the owner's passkey session
status: active
sourceFile: cmd/core/model_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T23:10:18.546Z'
id: TEST-core-model-service
type: test
---
Under task test:pg: a fresh core with the dev caller token, a passkey enrolled through core enroll-link and a software authenticator, then a login; the owner's session reads the model as off and empty, saves a 70/30 instrument and index model and reads it back on; a wrong caller token and an origin off the allow-list get permission_denied, an unknown session unauthenticated.