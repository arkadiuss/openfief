---
title: The Secret creator only ever asks the API server to create a Secret, and a Secret that exists answers 409 and is kept
status: active
sourceFile: internal/identity/kube/secrets_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-13T07:41:45.367Z'
id: TEST-identity-kube-secret-creator
type: test
---
Runs kube.SecretCreator and app.InitPlatformCredentials over the client-go fake clientset, which records every call. A fresh namespace gets exactly the two Secrets openfief-spicedb-key and openfief-db-spicedb in the openfief namespace, with the right type, labels and data, and the only calls are creates of secrets: never a get, list, watch, update, patch or delete. With both Secrets there already, as an owner fed them (D-29), the creates answer already-exists, nothing is changed, the owner's values are as they were and the outcomes say kept. A forbidden answer is returned as an error, not taken for a Secret that exists. A false pass would be a creator that read the Secret before creating it, which the recorded verbs rule out.