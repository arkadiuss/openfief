// Package valuation holds prices, FX rates, each naming its base, and valuations of
// holdings, as pure values and functions.
//
// Prices are end of day (D-21). Every amount keeps its original currency;
// conversion into the pack's reporting currency is an explicit step with a rate and a rounding mode
// (D-20). Bond values come from the Polish jurisdiction rules in
// internal/jurisdiction/pl. Every valuation and total carries its gross,
// net and liquidation values side by side, with no basis to choose (D-77).
package valuation
