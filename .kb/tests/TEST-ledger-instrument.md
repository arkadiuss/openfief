---
title: ISINs are checked, check digit included, and an instrument needs an ISIN and a venue symbol
status: active
sourceFile: internal/ledger/instrument_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:23:39.367Z'
id: TEST-ledger-instrument
type: test
verification_scope: unit
---
Table tests in package ledger_test parse a public ISIN example whose check sum runs over letters and three synthetic ISINs with valid check digits, and reject a wrong check digit, two swapped digits, lower case, a digit in the country code, a letter as check digit, too short, too long and empty input with ErrInvalidISIN; a rejected parse returns the zero ISIN. Instrument.Validate accepts an ISIN plus symbol and rejects either one missing with ErrInvalidInstrument. A false pass would be a parser that checks only the format: the wrong-check-digit and swapped-digit cases have the right shape.