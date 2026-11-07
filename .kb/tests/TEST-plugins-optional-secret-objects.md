---
title: An optional manifest Secret becomes an optional volume and a required one stays required
status: active
sourceFile: internal/plugins/operator/optional_secret_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T21:54:57.689Z'
id: TEST-plugins-optional-secret-objects
type: test
---
TestBuildObjectsOptionalSecret compiles a synthetic connector whose manifest asks for a required token Secret and an optional ca-bundle Secret, both granted, and builds its Kubernetes objects with BuildObjects. It compares the pod spec Secret volumes: the ca-bundle volume names sample--ca-bundle with optional true and the token volume names sample--token with optional unset. A false pass would need the compile step to drop the flag while BuildObjects invents it, which the exact comparison rules out.