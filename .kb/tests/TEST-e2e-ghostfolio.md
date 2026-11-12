---
title: The Ghostfolio view syncs the demo record into a real Ghostfolio 3.93.0 exactly, changes nothing on a second sync and reverts a manual edit
status: active
sourceFile: test/e2e/ghostfolio_test.go
verification_scope: end_to_end
origin:
  kind: agent
  recorded_at: '2026-11-08T08:54:14.141Z'
id: TEST-e2e-ghostfolio
type: test
proof_contract:
  version: kibi.proof-contract.v1
  integration: e2e-ghostfolio
  required_proofs:
    - symbol_id: SYM-e2e-ghostfolio-test-test-ghostfolio
      target: default
  success_policy: all_required_first_attempt
proof_bindings:
  - symbol_id: SYM-e2e-ghostfolio-test-test-ghostfolio
    target: default
    source_file: test/e2e/ghostfolio_test.go
    native_id: test/e2e/ghostfolio_test.go::TestGhostfolio
---
Runs in its own shard (task e2e:ghostfolio, CI job e2e-ghostfolio) on kind and k3d, after the same TestMain install of the demo stack as the main run. It deploys Ghostfolio 3.93.0, pinned by digest, in the namespace openfief-e2e-ghostfolio with a one-instance CloudNativePG database and Redis 7.2.17, and a Secret of random JWT key, access token salt and Redis password made for the run. It signs an anonymous user up (the first user of a fresh Ghostfolio), sets the base currency to PLN, and puts a fee on an asset of the owner's own and a cash balance in the account Ghostfolio made. It installs the Ghostfolio view with the openfief CLI built from the checkout, the user's security token in e2e-ghostfolio--security-token, Ghostfolio's Service as its in-cluster service (D-74), a 1h interval and no listings. Each sync it checks is the pass a new pod runs at start, read from the pod's JSON log. The first pass must report every projected owned account created, every projected activity imported and balance set, nothing deleted, and as left out exactly the transactions of listed instruments; the projection is computed from what the read API answers a script token, with the view's own Project. Then MakePlan between that projection and what Ghostfolio's API returns must be empty; every activity's comment must name a current transaction of its account; each treasury bond series' net BUY less SELL units on its MANUAL asset must equal the read API's holding; each owned account's latest balance must equal the read API's cash in its currency. A second pass must report no change and leave every activity and balance line, IDs included, as it was. After the owner adds a unit to one owned activity with PUT /api/v1/activities, comment kept, the next pass must report one import and one delete, the edited activity must be gone, one new activity added, the balances unchanged and the projection held again; the owner's own account must hold its activity and balance unchanged. A false pass would need the view and Ghostfolio to agree on a wrong projection that also matches the read API's holdings, cash and transaction keys.