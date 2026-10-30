// Package jurisdiction is the contract of a jurisdiction pack (GR-13, D-70):
// the pure rules of one country behind one interface, Pack, so another
// country is another implementation of it. The Polish pack is
// internal/jurisdiction/pl.
//
// A pack names the tax wrappers an account can have (D-76), the default
// reporting currency (D-62) and number format (D-113), and computes the
// three values of D-77 for each kind of product in each wrapper: gross,
// before any tax or fee; net, what the owner keeps at the planned exit
// after the taxes due then; and liquidation, the cash in hand if
// everything were sold or redeemed and withdrawn today. The kinds of
// product are a holding priced at market (Listed), a bond or deposit
// valued at its accrued value (Accrued), and cash.
//
// Besides the pack, the package holds what does not depend on a country:
// an account's fee rule for estimating a sale's commission (D-78) and the
// number format's separators. Nothing here does I/O or reads a clock.
package jurisdiction
