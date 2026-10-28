---
title: Decimal parses, computes exactly and rounds only by an explicit mode
status: active
sourceFile: internal/kernel/decimal_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T04:48:17.554Z'
id: TEST-kernel-decimal
type: test
---
Table tests parse decimal text (including a 39-digit value and rejected forms such as 1,5), check that 0.1 + 0.2 is exactly 0.3 and that multiplication keeps every digit, round ties with half-even, half-up and down, including negative ties, and reject an unknown rounding mode with ErrUnknownRoundingMode. Division always takes places and a mode: 2 / 3 to 2 places is 0.67 half-even and 0.66 down, ties such as 0.125 go to the even digit or away from zero as asked, signs of dividend and divisor are respected, and a zero divisor fails with ErrDivisionByZero. Comparison checks Cmp, Sign, IsZero and Equal across scales (1.50 equals 1.5). A false pass would be a comparison that ignores digits: expected values are parsed from text and compared with Decimal.Equal.