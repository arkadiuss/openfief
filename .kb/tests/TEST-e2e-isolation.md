---
title: 'On k3d and kind, plugins are isolated: the XTB parser has no egress, the proxy refuses hosts outside a grant and logs no request data, bypasses fail, and both sides of the core calls refuse bad tokens'
status: active
sourceFile: test/e2e/isolation_test.go
verification_scope: end_to_end
origin:
  kind: agent
  recorded_at: '2026-11-06T20:29:57.965Z'
id: TEST-e2e-isolation
type: test
proof_contract:
  version: kibi.proof-contract.v1
  integration: e2e
  required_proofs:
    - symbol_id: SYM-e2e-isolation-test-plugin-isolation
      target: default
  success_policy: all_required_first_attempt
proof_bindings:
  - symbol_id: SYM-e2e-isolation-test-plugin-isolation
    target: default
    source_file: test/e2e/isolation_test.go
    native_id: test/e2e/isolation_test.go::TestPluginIsolation
---
TestPluginIsolation installs, with a Plugin and a PluginGrant each, the XTB connector built from the checkout (xtb-iso, a manifest with no egress host, a grant approving its capability and institution) and the nbp-fx connector (egress-iso, granted api.nbp.pl and example.org, configured to ask NBP for nothing), both signed with the run key in the suite registry, and waits for both, the demo ui instance and the core. It then adds an ephemeral container running the client pods curl image to the XTB pod, the egress-iso pod and the core pod: it shares the pod network, so the pod NetworkPolicy applies to it exactly as to the plugin. It checks one pod per instance with the plugin container alone; the operator objects (the XTB NetworkPolicy allows egress only to the core pods on the plugin port, no HTTPS_PROXY, no egress token, an empty allow-list entry and no effective host, while egress-iso has the proxy rule, both tokens and api.nbp.pl:443); curl probes from the XTB pod (core plugin port 401; the proxy, the core browser port, another plugin and the API server own address get no connection; name lookups fail) with the outsider pod reaching the same API server address and resolving its name as the control; CONNECTs from egress-iso through the proxy, with a TokenRequest token of its ServiceAccount, to a host granted but not requested, an IP address and an in-cluster service, without a token, and with the token of a plugin granted nothing, each answered 407, and a plain HTTP request whose path, query, header and body carry a canary answered 407; direct connections and name lookups from egress-iso failing. It reads the egress proxy pod log until each refusal is logged as a CANONICAL-PROXY-DECISION line with the instance as role, the proxy type and the bare host, then checks every decision line of the run holds only known fields and a bare host:port, every allowed CONNECT is to a host on its instance allow-list, and no line holds the canary or a token. Finally it calls the XTB connector ImportBondLots from the core pod with the core token for the instance (501, the auth check passed) and with no token, a malformed one, the core token for egress-iso and the plugin own token (401), and from the client, outsider and egress-iso pods with the valid token (no connection); and the core plugin port from the XTB pod: read API and AuthService serve the ui view plugin token (200) and refuse no token, a malformed one, the XTB own core token and the ui egress token (401 and 403), and the browser port refuses a malformed and a view plugin token (401). Tokens go to curl on stdin and are never logged. A false pass would need a probe that cannot reach anything, which the 401 from the core plugin port, the 407 answers and the outsider controls rule out, or a proxy that logs nothing, which the wait for every expected line rules out.