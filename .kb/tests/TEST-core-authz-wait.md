---
title: A core waiting for SpiceDB says what is missing, by name, in /readyz and its log
status: active
sourceFile: cmd/core/authzwait_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-22T08:33:16.453Z'
id: TEST-core-authz-wait
type: test
---
Table tests of spicedbWaitReason give the exact text for each cause: no role spicedb and no database spicedb in the openfief-db cluster (install step 5 skipped), only one of them missing, a preshared key SpiceDB refuses (gRPC Unauthenticated), SpiceDB not answering while its role and database exist, and SpiceDB not answering while the database can't be asked. TestSpiceDBDiagnosis runs the diagnosis over a fake catalog row and wrapped gRPC errors, and checks it asks pg_catalog.pg_roles and pg_catalog.pg_database for spicedb. TestNotReadySaysWhatIsMissing wires a core whose SpiceDB refuses and whose database has neither, runs the start's reconcile and checks that /readyz answers 503 with the wait and the missing role and database by name, that the log line is a warning naming them and install step 5, and that the core is ready once SpiceDB answers. Under task test:pg, TestReadSpiceDBStoreOnPostgres runs the catalog query on a real PostgreSQL as an ordinary role and compares it with what the server holds. TestDemoFollowsTheLoader runs followDemo with a millisecond poll over a fake ledger and a fake reconcile whose every answer the test gives: nothing before the start's reconcile succeeded, one reconcile once it has, none while the accounts stay the same or the ledger is empty, one when the reseed's accounts appear, a failure logged as a warning and retried at the next poll, and a stop with its context. TestDemoSweepsWhateverItReads checks the sweep reconciles with nothing changed, and TestPlacementsIgnoreOrder that the comparison ignores order and sees a moved account. A false pass would be a message that is right only for one cause, which the table of causes and the exact strings rule out.