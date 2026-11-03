---
title: openfief plugin install stops before confirmation when its Secret is another instance's
status: active
sourceFile: internal/plugins/cli/secrets_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-03T08:22:07.498Z'
id: TEST-plugins-cli-shared-secret
type: test
---
TestRunRefusesSharedSecret runs the CLI against a fake cluster where instance sample-x names the Secret sample-key, and installs an image whose role key would name sample-key as instance sample: the run must fail with ErrSecretInUse, install nothing, and never print the confirmation prompt.