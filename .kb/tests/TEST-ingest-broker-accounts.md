---
title: Broker accounts validate and resolve into the OpenFief accounts they are mapped to
status: active
sourceFile: internal/ingest/brokeraccount_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T11:35:25.856Z'
id: TEST-ingest-broker-accounts
type: test
verification_scope: unit
---
Table tests of ingest.BrokerAccountID.Validate with synthetic Flex IDs, and of ResolveBrokerAccounts over a mapping with a direct transaction and entries of two synthetic broker accounts: nothing known leaves both unmapped and books only the direct one; a mapped account is booked with its account in each key source; an opening batch seeds a mapped account not yet seeded, leaves out the seeding of one already seeded but keeps its assertions, and seeds a mapped account the delivery doesn't name; the input is not changed. ReportedBrokerAccounts lists listed accounts then named ones.