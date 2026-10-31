---
title: Money keeps its currency, is exact, and rejects mixed currencies
status: active
sourceFile: internal/kernel/money_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T04:48:15.486Z'
id: TEST-kernel-money
type: test
verification_scope: unit
---
Table tests in package kernel_test call the public Money API with synthetic amounts. They check that add, subtract and negate stay in the operands' currency; that a price times a fractional quantity is exact and changes only when Round is called with a mode; that Add, Sub and Cmp on PLN and USD (both orders, and no currency against EUR) return a *CurrencyMismatchError naming both currencies; and that String, Neg, Mul and Round keep the original currency. A false pass would be an error that is not a *CurrencyMismatchError, or a result compared without its currency: go-cmp uses Money.Equal, which compares both.