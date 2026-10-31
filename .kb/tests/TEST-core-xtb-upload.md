---
title: An XTB upload through the real core waits for ticker confirmations and books them once confirmed
status: active
sourceFile: cmd/core/xtb_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-10-31T12:54:39.538Z'
id: TEST-core-xtb-upload
type: test
---
Runs the real core against a disposable PostgreSQL with a passkey owner session, the demo registry and an in-process connector that maps synthetic exports with xtb.Mapper and proposes an ISIN for TESTA.PL only. It uploads an export for an unmapped broker account (nothing booked), maps it, uploads again (deposit and withdrawal booked, four rows waiting), lists the pending tickers with the proposal, confirms one proposed and one typed ISIN (each books its rows) and checks the refusals, uploads again (all unchanged), reads the booked instruments through the read API, and searches every text column of the ingest and ledger schemas for the transfers' comments.