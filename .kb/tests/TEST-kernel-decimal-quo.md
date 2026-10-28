---
title: Decimal division rounds exactly once with an explicit mode
status: active
sourceFile: internal/kernel/quotient_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:11:39.271Z'
id: TEST-kernel-decimal-quo
type: test
---
A table test divides synthetic decimals and compares the quotient, parsed from text, with Decimal.Equal. It covers exact quotients, 2/3 with half-up and down, negative dividends and divisors (half-up goes away from zero, down goes toward zero), exact ties 1/8 and 3/8 under half-up and half-even, a value just above a tie, zero places, a bond-accrual quotient 535/365 = 1.4657 that must give 1.47, division by zero (ErrDivisionByZero) and an unknown rounding mode (ErrUnknownRoundingMode). A false pass would be double rounding through a float or a fixed intermediate precision: the tie cases would then come out wrong.