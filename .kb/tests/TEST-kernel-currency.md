---
title: ParseCurrency accepts only ISO 4217 alphabetic codes
status: active
sourceFile: internal/kernel/currency_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T04:48:21.341Z'
id: TEST-kernel-currency
type: test
verification_scope: unit
---
A table test parses PLN and USD, which must round-trip through String, and rejects lower case, two and four letters, the numeric code 985, the empty string and a non-ASCII code with ErrInvalidCurrency, compared with cmpopts.EquateErrors. A false pass would be accepting lower case or digits; both have explicit cases.