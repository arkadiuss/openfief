---
title: On k3d and kind, the operator holds a Plugin without a grant, refuses an image named by tag, holds an upgrade until its grant approves its new host, and calls only a catalog-pinned image verified
status: active
sourceFile: test/e2e/trust_test.go
verification_scope: end_to_end
origin:
  kind: agent
  recorded_at: '2026-11-06T19:43:37.678Z'
id: TEST-e2e-operator-trust
type: test
proof_contract:
  version: kibi.proof-contract.v1
  integration: e2e
  required_proofs:
    - symbol_id: SYM-e2e-trust-test-operator-requires-grant
      target: default
    - symbol_id: SYM-e2e-trust-test-operator-refuses-tag-reference
      target: default
    - symbol_id: SYM-e2e-trust-test-operator-upgrade-waits-for-grant
      target: default
    - symbol_id: SYM-e2e-trust-test-operator-verified-only-when-catalog-pinned
      target: default
  success_policy: all_required_first_attempt
---
TestOperatorRequiresGrant applies a Plugin of the suite-signed nbp-fx image without a PluginGrant and compares its status and objects once the signature is decided: Ready False NoGrant, everything requested pending, no Deployment, pod, allow-list entry or deployed version; the grant then makes it Ready. TestOperatorRefusesTagReference applies an image named by tag alone, which the API server must refuse as invalid, and one with a tag next to its digest, which the operator must refuse as InvalidImage and run nothing. TestOperatorUpgradeWaitsForGrant runs the synthetic connector, upgrades it to its image whose manifest asks for example.org, and checks that the old image keeps running, with the host pending and off the allow-list, until the grant approves it, after which the new image runs with the host on the allow-list. TestOperatorVerifiedOnlyWhenCatalogPinned copies the nbp-fx image the embedded catalog pins, with its CI signatures, from GHCR to the suite registry at the same digest and checks status.signature: Verified by the catalog publisher for it, Unverified by the owner signer for the checkout image. A false pass would need the operator to report a condition its objects contradict, which stateOf reads from the cluster as well.