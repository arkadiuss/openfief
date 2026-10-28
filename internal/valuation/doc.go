// Package valuation holds prices, FX rates to PLN and valuations of
// holdings, as pure values and functions.
//
// Prices are end of day (D-21). Every amount keeps its original currency;
// conversion to PLN is an explicit step with a rate and a rounding mode
// (D-20). Bond values come from the Polish jurisdiction rules in
// internal/jurisdiction/pl, and which of them counts in totals is the
// configured valuation basis, net-if-redeemed by default (D-25).
package valuation
