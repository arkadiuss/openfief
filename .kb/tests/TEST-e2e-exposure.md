---
title: On k3d and kind, the bundle install exposes nothing, the HTTPRoute component reaches the core and the UI over HTTPS through a Gateway, and with its auth-origin patch a passkey works on the route's hostname
status: active
sourceFile: test/e2e/exposure_test.go
verification_scope: end_to_end
origin:
  kind: agent
  recorded_at: '2026-11-06T19:43:39.007Z'
id: TEST-e2e-exposure
type: test
proof_contract:
  version: kibi.proof-contract.v1
  integration: e2e
  required_proofs:
    - symbol_id: SYM-e2e-exposure-test-bundle-exposure
      target: default
  success_policy: all_required_first_attempt
---
TestBundleExposure runs in the bundle run. It reads the bundle files and the openfief namespace and fails on any Ingress, Gateway API Gateway or route, DaemonSet, StatefulSet, Service that is not ClusterIP or has external IPs, pod on the host network or ports, or pod that belongs to none of the bundle Deployments, the plugin Deployments or the database. It then applies the route and NetworkPolicy of deploy/gateway with their placeholders replaced, creates a Gateway of the cluster implementation (Envoy Gateway on kind, Traefik on k3d) with a certificate from a CA made for the run, port-forwards to its data plane and, naming the test hostname in TLS and HTTP, expects /readyz 200, /v1/accounts 401 and the UI login page, each verified against that CA. Last, after every other bundle test has finished (routepasskey_test.go), it applies the component's auth-origin patch with the test hostname, waits for the core's rollout, runs openfief enroll --origin for the hostname and has Chromium, sent to the Gateway and trusting only the run's CA and Gateway keys, create a passkey there, sign out and sign in again.