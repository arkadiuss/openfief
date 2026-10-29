// Package pl holds Polish jurisdiction rules as pure functions: retail
// treasury bond series and their statutory valuation, tax wrappers (IKE,
// IKZE), Belka tax on bond interest, which NBP rate tax uses, and the
// business days that pick the NBP reference rate for ROR and DOR.
//
// Valuation uses these rules; nothing here depends on valuation or ledger.
// Series terms that change with every issue (first-period rate, margin,
// early-redemption fee) and the published indices (CPI from GUS, the NBP
// reference rate) are inputs, never hard-coded.
//
// Sources:
//   - Issue letters (listy emisyjne) of each series, linked as "Zobacz list
//     emisyjny" from its offer page under
//     https://www.obligacjeskarbowe.pl/oferta-obligacji/ and listed at
//     https://www.obligacjeskarbowe.pl/listy-emisyjne/. Doc comments cite the
//     October 2026 letters: OTS0127 (nr 94/2026), ROR1027 (nr 95/2026),
//     DOR1028 (nr 96/2026), TOS1029 (nr 97/2026), COI1030 (nr 98/2026),
//     EDO1036 (nr 99/2026), ROS1032 (nr 100/2026) and ROD1038 (nr 101/2026).
//   - Interest tables (tabele odsetkowe) with each series' daily accrued
//     interest per bond: https://www.obligacjeskarbowe.pl/tabela-odsetkowa/.
//   - Personal income tax act (ustawa o podatku dochodowym od osób
//     fizycznych): art. 11a ust. 1 (NBP average rate of the last business day
//     before the day of income), art. 30a ust. 1 pkt 2 (19% on interest and
//     discount of securities), art. 21 ust. 1 pkt 58 (IKE exemption):
//     https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU19910800350
//   - Act on days off work (ustawa o dniach wolnych od pracy), art. 1, for
//     the business days that place the reference rate's day (IsBusinessDay):
//     https://api.sejm.gov.pl/eli/acts/DU/1951/28
//   - Tax ordinance (Ordynacja podatkowa) art. 63 § 1a (rounding of the tax
//     on interest; see BelkaTax for how it differs from the issuer's examples):
//     https://isap.sejm.gov.pl/isap.nsf/DocDetails.xsp?id=WDU19971370926
package pl
