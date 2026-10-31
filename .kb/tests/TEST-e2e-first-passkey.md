---
title: End to end, openfief enroll and a browser's virtual authenticator create and use the first passkey on localhost
status: active
sourceFile: test/e2e/passkey_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T12:01:54.870Z'
id: TEST-e2e-first-passkey
type: test
verification_scope: end_to_end
proof_contract:
  version: kibi.proof-contract.v1
  integration: e2e
  required_proofs:
    - symbol_id: SYM-fd-e2e-test-first-passkey
      target: default
  success_policy: all_required_first_attempt
proof_bindings:
  - symbol_id: SYM-fd-e2e-test-first-passkey
    target: default
    source_file: test/e2e/passkey_test.go
    native_id: test/e2e/passkey_test.go::TestFirstPasskey
---
On kind and k3d with the operator-deployed UI: openfief enroll execs core enroll-link in the core's pod and forwards localhost:8080 to the core; /readyz answers through the forward; Chromium driven by passkey.mjs with the WebAuthn virtual authenticator opens the link, creates the passkey, lands on the overview signed in, keeps __Host-openfief_session Secure, HttpOnly, SameSite=Strict on http://localhost, signs out and signs in again with the passkey. No flag weakens the origin check.