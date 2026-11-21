---
title: End to end, an upgrade in place from the published 2026.11.0-alpha.1 to the checkout's bundle keeps every stored record and the D-209, D-212 and D-216 behaviors
status: active
sourceFile: test/e2e/upgrade_test.go
verification_scope: end_to_end
origin:
  kind: agent
  ref: the design session's exit criterion (a) of 2026.11.0-alpha.2, relayed by the coordinator, 2026-11-21
  recorded_at: '2026-11-21T12:00:00.000Z'
id: TEST-e2e-upgrade-keeps-stored-data
type: test
proof_contract:
  version: kibi.proof-contract.v1
  integration: e2e
  required_proofs:
    - symbol_id: SYM-e2e-upgrade-test-test-upgrade-from-published
      target: default
  success_policy: all_required_first_attempt
---
A Go test with the e2e build tag in test/e2e/upgrade_test.go, run in task e2e's upgrade run (OPENFIEF_E2E_UPGRADE). Its TestMain installs the published release that scripts/release-fetch.sh downloaded and cosign-verified, with kubectl apply --server-side as the install guide says; the test seeds it with the demo dataset through that release's own demo-seed image, upgrades it to the release candidate's bundle with the guide's Upgrade steps, and compares a content hash of every row of every table the contexts own before and after, reads the grants, the cutover record and the audit log from the database, the Secrets from the cluster, and the readiness, the read API, the MCP endpoint and the UI's login page through a port-forward. A false pass would be a seed that stored nothing, which the test rules out by requiring rows in every kind of record first; an upgrade that reseeds instead of migrating, which the unchanged rows rule out; or a cutover that runs again on a restart, which the test restarts the core to catch.