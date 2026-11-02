---
title: End to end, the CLI built from the checkout installs the UI and enrolls the first passkey on a bundle install
status: active
verification_scope: end_to_end
sourceFile: test/e2e/firstrun_test.go
origin:
  kind: agent
  recorded_at: '2026-11-01T21:52:02.106Z'
id: TEST-e2e-bundle-first-run
type: test
proof_contract:
  version: kibi.proof-contract.v1
  integration: e2e
  required_proofs:
    - symbol_id: SYM-e2e-firstrun-test-bundle-first-run
      target: default
  success_policy: all_required_first_attempt
---
A Go test with the e2e build tag in test/e2e/firstrun_test.go, run in task e2e's bundle run. Its TestMain installs the release bundle rendered from this checkout's images, CloudNativePG and the openfief-db cluster from the bundle included, signs the plugin images in the suite's registry with a key made for the run and trusts it, and builds the openfief CLI from the checkout with go build, reading through an overlay a catalog that pins the UI's digest in the suite's registry to that key as its publisher, as a release does with its own catalog. The test runs that binary as its own process: `openfief plugin install ui`, answered y, must print its question and Installed ui., and the operator must run the UI with the browser path /; then `openfief enroll` must print the one-time link for http://localhost:8080, with which Chromium and its virtual authenticator, through passkey.mjs, create the first passkey, keep the __Host- cookie, sign out and sign in again; Ctrl-C must end enroll with exit status 0. A false pass would be an install or enrolment that bypasses the CLI, which running the built binary rules out.