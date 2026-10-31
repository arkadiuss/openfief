---
title: A connector's mapping, FX rates and mode are written as connector contract messages
status: active
sourceFile: internal/ingest/delivery/delivery_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T17:06:38.916Z'
id: TEST-ingest-delivery
type: test
verification_scope: unit
---
Table tests in package delivery_test build a synthetic mapping with an opening lot, a fee, holding and cash assertions and both problem kinds, and compare FromMapping's Delivery field by field with protocmp: decimal strings without trailing zeros or exponents, YYYY-MM-DD dates, a lot-opened date only where set, a quantity only on holding postings and the instrument's ISIN and symbol. Zero transaction, ledger account and problem kinds are refused with ErrUnknownValue. The ibkrflex adapter's synthetic activity statement is mapped in both modes and every transaction, assertion and problem must be written in order with the same amounts. FXRates and Mode are checked too. A false pass would be a lossy decimal; the 0.0001 and trailing-zero cases guard it.