---
title: The core hands an owner's bond lot to the account's one bound connector and imports the delivery
status: active
sourceFile: internal/ingest/app/bondlots_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T01:13:42.962Z'
id: TEST-ingest-app-bondlots
type: test
---
Use-case tests of BondLots.Add over in-memory fakes: the owner's lot goes to the one instance bound to the account as batch b-1; the delivery is imported once as a transactions batch with source entry, its keys scoped to plugin:bonds-main, and the series' terms stored. Refusals, each reaching neither ledger nor stores: no session, a read-scoped session, a lot after today or under the wrong series, no or two bound instances, unreadable bindings, the connector refusing or not taking bond lots, and a delivery into another account (ErrAccountNotBound).