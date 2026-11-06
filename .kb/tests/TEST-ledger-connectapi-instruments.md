---
title: The InstrumentService renames for the owner's view and maps refusals
status: active
sourceFile: internal/ledger/connectapi/instruments_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-06T06:58:53.013Z'
id: TEST-ledger-connectapi-instruments
type: test
---
Serves InstrumentService over httptest with a fake store and owner check: the owner's view gets the cleaned name back and stores an owner's record; a read session is permission_denied, no session unauthenticated, another caller permission_denied, a malformed ISIN or empty name invalid_argument, a failing store internal with no text.