// Package pl holds Polish jurisdiction rules as pure functions: retail
// treasury bond series and their statutory valuation, tax wrappers (IKE,
// IKZE), Belka tax on bond interest, and which NBP rate tax uses.
//
// Valuation uses these rules; nothing here depends on valuation or ledger.
// Series terms that change with every issue (first-period rate, margin,
// early-redemption fee) and the published indices (CPI from GUS, the NBP
// reference rate) are inputs, never hard-coded.
//
// Sources:
//   - Offer and issue letters (listy emisyjne) of each series:
//     https://www.obligacjeskarbowe.pl/oferta-obligacji/ and
//     https://www.obligacjeskarbowe.pl/listy-emisyjne/
//   - Personal income tax act (ustawa o podatku dochodowym od osób
//     fizycznych): art. 11a ust. 1 (NBP average rate of the last business day
//     before the day of income), art. 30a ust. 1 pkt 2 (19% on interest),
//     art. 21 ust. 1 pkt 58 (IKE exemption):
//     https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU19910800350
//   - Tax ordinance (Ordynacja podatkowa) art. 63 § 1a (withheld tax on
//     interest rounded to whole grosze):
//     https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU19971370926
package pl
