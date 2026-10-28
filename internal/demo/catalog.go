package demo

import (
	"fmt"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

// The demo's sleeves and accounts. Their IDs say "demo", so that no one can
// take them for a real portfolio.
const (
	DefensiveSleeve portfolio.SleeveID = "demo-defensive"
	OffensiveSleeve portfolio.SleeveID = "demo-offensive"

	IBKRAccount  ledger.AccountID = "demo-ibkr"
	XTBAccount   ledger.AccountID = "demo-xtb"
	BondsAccount ledger.AccountID = "demo-bonds-ike"

	// IBKRFlexAccount is the IBKR account number in the demo's Flex
	// statements: the all-zero placeholder the ibkrflex tests use too.
	IBKRFlexAccount = "U0000000"

	// PriceSource names the source of every demo price.
	PriceSource = "demo-synthetic"
)

// The institutions, as the ingest context names them.
const (
	IBKRInstitution  = "ibkr"
	XTBInstitution   = "xtb"
	BondsInstitution = "treasury-bonds"
)

// The venue symbols of the demo's market-priced instruments.
const (
	symWorld    = "DWLD"
	symTech     = "DTEC"
	symEuroDiv  = "DEUD"
	symBank     = "DPLB"
	symEnergy   = "DPLE"
	symEmerging = "DEMG"
)

// The ISO codes of the demo's currencies.
const (
	codeUSD = "USD"
	codeEUR = "EUR"
	codePLN = "PLN"
)

// The key sources of the transactions the demo books through the ledger's
// API: one per account, as connectors key theirs (ledger.TransactionKey).
const (
	xtbSource   = "demo-xtb:" + string(XTBAccount)
	bondsSource = "demo-bonds:" + string(BondsAccount)
)

// Instrument is one synthetic instrument of the demo, traded at a market
// price, with what its price path starts from.
type Instrument struct {
	Instrument ledger.Instrument
	Name       string
	Currency   kernel.Currency
	// startCents is the first close, in hundredths of Currency.
	startCents int64
	// driftPPM and volPPM are the mean and the half-range of the daily
	// return, in millionths.
	driftPPM, volPPM int64
}

// instrumentSpec is the source of an Instrument.
type instrumentSpec struct {
	country  string
	number   int
	symbol   string
	name     string
	currency string
	start    int64
	drift    int64
	vol      int64
}

// instrumentSpecs are the demo's market-priced instruments: three at IBKR,
// three at XTB. The numbers keep their synthetic ISINs apart.
func instrumentSpecs() []instrumentSpec {
	return []instrumentSpec{
		{"IE", 1, symWorld, "Demo World ETF", codeUSD, 10_000, 350, 9_000},
		{"US", 2, symTech, "Demo Tech Co", codeUSD, 18_000, 600, 18_000},
		{"DE", 3, symEuroDiv, "Demo Euro Dividend Co", codeEUR, 4_500, 150, 8_000},
		{"PL", 4, symBank, "Demo Polska Bank", codePLN, 6_000, 300, 14_000},
		{"PL", 5, symEnergy, "Demo Polska Energia", codePLN, 2_500, 100, 16_000},
		{"IE", 6, symEmerging, "Demo Emerging Markets ETF", codePLN, 3_200, 250, 10_000},
	}
}

// Instruments returns the demo's market-priced instruments, in the order of
// instrumentSpecs.
func Instruments() ([]Instrument, error) {
	specs := instrumentSpecs()
	out := make([]Instrument, 0, len(specs))
	for _, s := range specs {
		isin, err := SyntheticISIN(s.country, s.number)
		if err != nil {
			return nil, err
		}
		c, err := kernel.ParseCurrency(s.currency)
		if err != nil {
			return nil, fmt.Errorf("demo instrument %s: %w", s.symbol, err)
		}
		out = append(out, Instrument{
			Instrument: ledger.Instrument{ISIN: isin, Symbol: s.symbol}, Name: s.name, Currency: c,
			startCents: s.start, driftPPM: s.drift, volPPM: s.vol,
		})
	}
	return out, nil
}

// Sleeves returns the demo's two sleeves: defensive for the bonds and their
// cash, offensive for the brokerage accounts.
func Sleeves() []portfolio.Sleeve {
	return []portfolio.Sleeve{
		{ID: DefensiveSleeve, Name: "Demo defensive", Role: portfolio.RoleDefensive},
		{ID: OffensiveSleeve, Name: "Demo offensive", Role: portfolio.RoleOffensive},
	}
}

// Accounts returns the demo's three accounts.
func Accounts() []ledger.Account {
	return []ledger.Account{
		{ID: IBKRAccount, Institution: IBKRInstitution, Name: "Demo IBKR", Sleeve: OffensiveSleeve},
		{ID: XTBAccount, Institution: XTBInstitution, Name: "Demo XTB", Sleeve: OffensiveSleeve},
		{ID: BondsAccount, Institution: BondsInstitution, Name: "Demo treasury bonds (IKE)", Sleeve: DefensiveSleeve},
	}
}
