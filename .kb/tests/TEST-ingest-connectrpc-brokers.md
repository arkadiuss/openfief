---
title: The core translates entries a connector names by broker account apart from the others
status: active
sourceFile: internal/ingest/connectrpc/brokers_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T11:35:30.855Z'
id: TEST-ingest-connectrpc-brokers
type: test
---
Test of deliveryFrom over a synthetic delivery whose transaction and balance assertion name the broker account U0000002 and whose other transaction names acc-1, with two listed broker accounts: the acc-1 transaction stays direct, U0000002's entries are grouped apart in no account with their key kept, and both listed accounts are reported. Refusals of entries naming both accounts and of malformed broker accounts are in TestDeliveryFromRejects.