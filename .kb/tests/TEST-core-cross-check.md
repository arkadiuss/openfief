---
title: The core hands a delivery's cross-check prices to valuation's cross-check use cases
status: active
sourceFile: cmd/core/crosscheck_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T06:24:24.737Z'
id: TEST-core-cross-check
type: test
verification_scope: unit
---
TestCrossCheckInputs builds valuation's CrossChecks over an in-memory store, the way cmd/core wires it, and calls crossCheckInputs.StoreCrossCheckPrices, the ingest port Connectors uses: a synthetic IBKR mark of PLTEST000010/TSTA from instance ibkr-main is stored unchanged, and a batch with a zero close is refused with ErrInvalidInput. A false pass would need the port to bypass the use case's validation, which the zero-close case rules out.