// Package kernel is the shared kernel of the bounded contexts: exact decimal
// numbers, currencies, money, quantities and calendar dates.
//
// Everything here is a pure value type. Arithmetic is exact; rounding happens
// only where a caller asks for it, with an explicit RoundingMode. Money never
// mixes currencies: an operation on two different currencies returns a
// *CurrencyMismatchError instead of a result.
package kernel
