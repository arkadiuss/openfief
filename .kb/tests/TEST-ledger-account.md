---
title: Accounts belong to one sleeve and ledger accounts fit their kind
status: active
sourceFile: internal/ledger/account_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:23:42.098Z'
id: TEST-ledger-account
type: test
verification_scope: unit
---
Table tests validate a synthetic account and reject one with no sleeve, ID, institution or name with ErrInvalidAccount. Ledger account (Book) tests cover all six kinds: cash, fee, withholding tax, FX and equity need a currency and no instrument, a holding needs a valid instrument and no currency, and the zero kind is rejected, all with ErrInvalidBook; String names the kind and its currency or instrument. The constructors CashBook ... EquityBook are compared field by field with go-cmp and must validate. A false pass would be a holding that also carries a currency; it has its own case.