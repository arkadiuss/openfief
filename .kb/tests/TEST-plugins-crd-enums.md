---
title: Every Enum of the operator's CRDs lists exactly the values Go accepts
status: active
sourceFile: internal/plugins/operator/crd_enums_test.go
verification_scope: unit
origin:
  kind: agent
  ref: 'Design session, on the author''s behalf: CRD enum drift test after 0a9c935, 2026-11-18'
  recorded_at: '2026-11-18T08:12:40.601Z'
id: TEST-plugins-crd-enums
type: test
---
A table test in package operator_test reads the generated CRDs in deploy/crds and, for each field path a kubebuilder Enum marker covers (the capabilities of Plugin spec, deployed, effective and pending approval and of PluginGrant spec; signature trust and signerKind), compares the enum with the Go set the parser or trust policy uses: plugins.Capabilities, plugins.Trusts, plugins.SignerKinds. A value Go accepts that a CRD lacks makes the API server refuse a valid Plugin, as model-provider did before 0a9c935. A false pass would be a path that silently resolves to no enum; a missing enum fails the test rather than comparing empty sets.