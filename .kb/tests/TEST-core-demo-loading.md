---
title: The core refuses the first-run wizard's writes while the demo data loads
status: active
sourceFile: cmd/core/demoloading_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-03T19:03:57.491Z'
id: TEST-core-demo-loading
type: test
---
TestRefuseWhileDemoLoads serves the generated SleeveService and AccountService handlers behind refuseWhileDemoLoads on an httptest server and calls them with the generated Connect clients: while the demo is loading, AddSleeve and CreateAccount fail with FAILED_PRECONDITION and ListBrokerAccounts is served; on a ready demo or an install without demo sessions all are served; a demo marker that can't be read refuses the writes as INTERNAL.