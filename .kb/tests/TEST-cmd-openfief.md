---
title: The openfief binary verifies real cosign bundles with the operator's verifier before installing
status: active
sourceFile: cmd/openfief/main_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T23:35:59.866Z'
id: TEST-cmd-openfief
type: test
---
Tests in package main build the real Env (embedded catalog, cosign.Verifier on the bundled trust root, operator.ValidateConfig, kube.ResourcesYAML) and replace only the cluster with controller-runtime's fake client. cosigntest serves an in-memory registry with an image signed by a fresh key as `cosign sign --key` does and an unsigned one. Installing the signed image works with the key from a --trusted-signers file or from the cluster's openfief-trusted-signers ConfigMap, and creates a Plugin and a grant of exactly the signed label; without the key, unsigned, or with a config the schema rejects it is refused. --print changes nothing; install, list and uninstall then round-trip. A false pass would be an image accepted without a trusted key; that case runs with no signers at all.