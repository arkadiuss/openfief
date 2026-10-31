---
title: A connector reads bond lot messages exactly and writes bond terms as contract messages
status: active
sourceFile: internal/ingest/delivery/bonds_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T01:13:46.830Z'
id: TEST-ingest-delivery-bonds
type: test
verification_scope: unit
---
TestBondLots reads a synthetic EDO0536 lot message into ingest.BondLot with exact decimals; TestBondLotsRefuses checks a missing or unknown series, non-numeric quantity, rate, margin or fee, a non-ISO date and a missing fee or fee currency fail with ErrMalformedLot; TestBondTerms writes terms as rates and a PLN fee in plain decimal strings. A false pass would need a parse that accepts text the contract forbids.