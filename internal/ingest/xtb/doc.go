// Package xtb maps XTB's xStation "Historia operacji" export, an .xlsx file
// of one account (R-24), into the connector contract's mapping (D-9). It
// reads the workbook itself, with the standard library and bounded sizes,
// and never converts an amount through a float: a number cell's decimal
// string is parsed exactly and rounded to the currency's minor unit.
//
// The export names instruments by ticker and market suffix, with no ISIN,
// so stock and ETF rows become ticker entries that the core books once the
// owner confirms each ticker's ISIN (D-116). CFDs are flagged (D-24).
package xtb
