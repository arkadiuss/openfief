---
title: On k3d and kind, openfief plugin install refuses the images the operator refuses, with the operator's reason, and applies nothing
status: active
sourceFile: test/e2e/refusal_cli_test.go
verification_scope: end_to_end
proof_contract:
  version: kibi.proof-contract.v1
  integration: e2e
  required_proofs:
    - symbol_id: SYM-e2e-refusal-cli-test-cli-refusal-agrees-with-operator
      target: default
  success_policy: all_required_first_attempt
proof_bindings:
  - symbol_id: SYM-e2e-refusal-cli-test-cli-refusal-agrees-with-operator
    target: default
    native_id: test/e2e/refusal_cli_test.go::TestCLIRefusalAgreesWithOperator
    source_file: test/e2e/refusal_cli_test.go
origin:
  kind: agent
  recorded_at: '2026-11-12T19:56:27.770Z'
id: TEST-e2e-cli-refusal
type: test
---
Runs on a local k3d or kind cluster with the demo's stack, the real operator, the suite's registry and the openfief CLI built from the checkout. For four images, an unsigned one, one signed with a key made for the test that the owner does not trust (its signature first verified valid), one named by tag alone and one naming a tag next to its digest, it runs openfief plugin install as a process and requires a non-zero exit, no Plugin and no PluginGrant, and the refusal text of the matching plugins error in its output. It then applies the same image with plain kubectl semantics, a grant approving everything requested plus a Plugin, and requires the operator's SignatureVerified False with reason Unsigned, UntrustedSigner or InvalidImage, a message holding the same refusal text (and the untrusted key's fingerprint), Ready False, no deployed version, no Deployment and no pod; for the tag alone it requires the API server to refuse the Plugin as invalid. A false pass would be a CLI that failed for another cause, such as no cluster access, which the refusal text it must print rules out, or an operator refusal for another reason, which the asserted reason and message rule out.