// Package treasurybonds is the Polish retail treasury bond connector's
// logic (D-9, D-49): it maps the bond lots the owner enters by hand in the
// built-in UI, which the core hands to the connector, into ledger
// transactions and the series terms valuation needs. Neither the registry
// nor the bank that sells the bonds offers an API (R-5), so manual entry is
// this connector's ingest path; the plugin plugins/treasury-bonds serves it
// to the core.
//
// A lot is what D-9 names (series, quantity, purchase date) plus its
// series' terms. The owner may leave the terms to the connector, which
// reads the first-period rate and margin from the Ministry of Finance's
// published data and the early-redemption fee from the reviewed table of
// the issue letters (D-97, CompleteTerms); terms the owner enters override
// them, and the latest entry wins. Each lot becomes a deposit of its
// cost and the buy of its bonds, on its purchase date, in the one ledger
// account the connector instance is configured for. The bonds' value comes
// from the valuation context's bond rules (jurisdiction/pl), never from a
// price.
//
// MinistryClient reads the ministry's published data (D-97): the
// spreadsheet "Dane dotyczące obligacji detalicznych", which lists every
// retail series since 1992 with each period's rate and its margin, from its
// gov.pl page, https://www.gov.pl/web/finanse/obligacje-detaliczne1, or,
// when that fails, from the same dataset on dane.gov.pl (dataset 805), with
// the attribution the ministry's re-use terms ask for (CC BY 4.0, R-27). It
// finds the file's link on every fetch, since the ministry replaces the file
// under a new address, and never reads obligacjeskarbowe.pl (D-8). The
// spreadsheet is an Excel 97-2003 workbook (.xls), which ParseWorkbook reads
// itself (ministry_cfb.go, ministry_biff.go, ministry_sst.go), keeping only
// text and number cells, within limits on its size, cells and strings; the
// ministry's layout never leaves the package. The reader lives here, not in
// valuation, because a plugin may import no new internal package
// (docs/dev/code-organization.md, "Plugins").
package treasurybonds
