---
title: On k3d and kind, SpiceDB runs from the install with generated credentials, and with the core on SpiceDB a stopped SpiceDB makes the read API refuse and a started one makes it answer again
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T04:07:15.947Z'
sourceFile: test/e2e/spicedb_test.go
verification_scope: end_to_end
proof_contract:
  version: kibi.proof-contract.v1
  integration: e2e
  required_proofs:
    - symbol_id: SYM-e2e-spicedb-test-spicedb-deployed
      target: default
    - symbol_id: SYM-e2e-spicedb-test-spicedb-fails-closed
      target: default
  success_policy: all_required_first_attempt
proof_bindings:
  - symbol_id: SYM-e2e-spicedb-test-spicedb-deployed
    target: default
    native_id: test/e2e/spicedb_test.go::TestSpiceDBDeployed
    source_file: test/e2e/spicedb_test.go
  - symbol_id: SYM-e2e-spicedb-test-spicedb-fails-closed
    target: default
    native_id: test/e2e/spicedb_test.go::TestSpiceDBFailsClosed
    source_file: test/e2e/spicedb_test.go
id: TEST-e2e-spicedb
type: test
---
TestSpiceDBDeployed waits for the Deployment openfief-spicedb of the demo's app layer (deploy/spicedb: one pod, an init container running spicedb migrate head, serve with the postgres engine) and for the Job openfief-spicedb-secrets to complete, then reads the Secrets the Job made: the preshared key and the role's password are 32 bytes or more and differ, and the username is spicedb. It waits until CloudNativePG applied the Database spicedb of the openfief-db cluster and lists the role spicedb as reconciled, and checks the pod spec: a tag and a digest on both images, non-root, a read-only root filesystem without privilege escalation or capabilities, a memory request and limit, no ServiceAccount token, and the key and the password from Secrets, never literals. It reads the logs of SpiceDB and of the core and fails if either holds the value of either Secret. TestSpiceDBFailsClosed is serial: with the policy the demo's script token lists the three demo accounts; the test adds --authz=spicedb and --spicedb-insecure to the core's arguments, waits for the log line that SpiceDB is reconciled (the schema and the relationships the core owns written), and the same token, which holds no grant, lists none; it scales SpiceDB to zero and waits for the read API to answer 5xx, never 200, with the core's /healthz still 200; it scales SpiceDB back, waits for it, and the read API answers again from the relationships PostgreSQL kept; and it returns the core to the policy and the token lists the three accounts again, also from its cleanup if a step failed. A false pass would be a token that listed nothing because the core never reached SpiceDB, which the reconciled log line and the policy's three accounts before and after rule out, or a refusal that was the core's crash, which the healthz answer rules out. The owner's own refusal and view of every account are the contract test's (task test:authz).