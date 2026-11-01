---
title: The operator refuses an unsigned plugin image and one signed by an untrusted key on k3d and kind
status: active
sourceFile: test/e2e/refusal_test.go
verification_scope: end_to_end
origin:
  kind: agent
  recorded_at: '2026-11-01T13:50:48.983Z'
id: TEST-e2e-plugin-refusals
type: test
proof_contract:
  version: kibi.proof-contract.v1
  integration: e2e
  required_proofs:
    - symbol_id: SYM-e2e-test-operator-refuses-untrusted-images
      target: default
  success_policy: all_required_first_attempt
proof_bindings:
  - symbol_id: SYM-e2e-test-operator-refuses-untrusted-images
    target: default
    source_file: test/e2e/refusal_test.go
    native_id: test/e2e/refusal_test.go::TestOperatorRefusesUntrustedImages
---
Runs on a local k3d or kind cluster with the demo's stack, the real operator and the suite's registry. The test pushes two synthetic images carrying the NBP FX connector's manifest label to that registry: one nobody signs, and one signed with an ECDSA key made in memory for the test, which is never added to the owner's trusted signers. Before applying it, the test verifies with the operator's own cosign verifier, trusting that key alone, that its signature is valid over the digest and the label, so only the signer can make the operator refuse it. For each image it applies a PluginGrant that approves everything the manifest requests and a Plugin, waits with a deadline for SignatureVerified of the current generation, and requires False with reason Unsigned for the first and UntrustedSigner for the second, whose message and status.untrustedSigners name the key as public key sha256 and its fingerprint (D-169); for both Ready False with SignatureNotVerified, ApprovalRequired False with Approved, no Deployment, no pod, no allow-list entry, no deployed version or signer in the status, and a Warning event on the Plugin with the condition's reason naming the refusal. A True condition fails at once. A false pass would be a refusal for another cause, which the grant and the asserted reasons rule out, or a signature never attached, which the suite's own verification rules out.