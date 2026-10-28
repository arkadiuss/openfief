// Package demo generates the demo dataset: a synthetic portfolio, about a
// year of its history and its valuation inputs, for showing OpenFief's real
// pipelines working on fake data (GR-12). cmd/demo-seed loads it.
//
// Everything here is a pure function of a seed and an injected date: the same
// seed and date always give the same dataset. Nothing in it is personal or
// account data. Instruments have synthetic ISINs with a valid check digit and
// "TESTD" after the country code (IsSynthetic); prices are seeded random
// walks; accounts, IDs and amounts are made up. The only real data is public:
// the Polish treasury bond series codes with their published issue terms, and
// GUS consumer price indices (bonds.go cites the sources). FX rates are not
// generated: the loader takes NBP table A rates, which are public too.
//
// The IBKR account is loaded through its real ingest path, so its history is
// written as synthetic Flex Query statements (flex.go) that the ibkrflex
// connector maps; XTB and the bond registry have no connector yet, so their
// history is ledger transactions for the ledger's own API.
package demo
