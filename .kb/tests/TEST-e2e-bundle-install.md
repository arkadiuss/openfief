---
title: OpenFief installs from the release bundle of this checkout's images and serves the read API
status: active
sourceFile: test/e2e/bundle_test.go
verification_scope: end_to_end
origin:
  kind: agent
  recorded_at: '2026-11-01T13:50:24.281Z'
id: TEST-e2e-bundle-install
type: test
proof_contract:
  version: kibi.proof-contract.v1
  integration: e2e
  required_proofs:
    - symbol_id: SYM-e2e-bundle-test-bundle-pins-image-digests
      target: default
    - symbol_id: SYM-e2e-bundle-test-bundle-becomes-ready
      target: default
    - symbol_id: SYM-e2e-bundle-test-bundle-read-api
      target: default
  success_policy: all_required_first_attempt
proof_bindings:
  - symbol_id: SYM-e2e-bundle-test-bundle-pins-image-digests
    target: default
    source_file: test/e2e/bundle_test.go
    native_id: test/e2e/bundle_test.go::TestBundlePinsImageDigests
  - symbol_id: SYM-e2e-bundle-test-bundle-becomes-ready
    target: default
    source_file: test/e2e/bundle_test.go
    native_id: test/e2e/bundle_test.go::TestBundleBecomesReady
  - symbol_id: SYM-e2e-bundle-test-bundle-read-api
    target: default
    source_file: test/e2e/bundle_test.go
    native_id: test/e2e/bundle_test.go::TestBundleReadAPI
---
Go tests with the e2e build tag in test/e2e/bundle_test.go, run by task e2e before the demo run with OPENFIEF_E2E_BUNDLE set. task e2e:images pushes this checkout's core, operator and egress-proxy images to the suite's registry and task e2e renders the release bundle from them with scripts/release-bundle.sh, each pinned by digest, with CloudNativePG's pinned manifest and the openfief-db cluster. TestMain removes any earlier OpenFief namespace, applies the bundle's CloudNativePG manifest, its CRDs file, the suite's trusted signer, its openfief-db cluster and its install file, all unchanged, then removes the namespace again after the tests. The tests check that every Deployment in the bundle runs its images by digest from the suite's registry and the cluster runs exactly those; that the operator, the egress proxy and the core become available; and, through a port-forward to the core as openfief enroll opens one, that /readyz answers 200, that a script token issued by the core's own CLI lists accounts with 200, and that a call without a token gets 401. A false pass would be an install that isn't the bundle, so the files are applied unedited and the live Deployments are compared with them.