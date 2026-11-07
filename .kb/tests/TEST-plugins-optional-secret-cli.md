---
title: The install summary marks an optional Secret as one the plugin runs without
status: active
sourceFile: internal/plugins/cli/optional_secret_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T21:54:58.429Z'
id: TEST-plugins-optional-secret-cli
type: test
---
TestRenderProposalOptionalSecret renders the install proposal of a synthetic connector with a required api-token Secret and an optional ca-bundle Secret, and checks that the ca-bundle line ends with optional: the plugin runs without it while the api-token line carries no such mark, so the owner sees which Secrets they must create before the plugin starts.