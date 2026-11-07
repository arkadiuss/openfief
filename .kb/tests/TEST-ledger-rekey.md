---
title: A re-keyed transaction moves only its holding postings of the provisional instrument; display names carry over once
status: active
sourceFile: internal/ledger/rekey_test.go
origin:
  kind: agent
  recorded_at: '2026-11-07T22:44:16.797Z'
id: TEST-ledger-rekey
type: test
---
Table tests of ledger.Rekeyed over a buy, a coupon booked as interest, another instrument's buy and a deletion: only postings to the provisional instrument move, the original is untouched, the result is valid and its fingerprint differs only when something moved, so importing it supersedes; and of RekeyedNames carrying the first delivered name and the owner's latest rename to the registry ISIN once, never over a name of the same kind it has.