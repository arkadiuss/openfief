package demo

import (
	"fmt"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/portfolio"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// historyDays is how long before today the demo goes live: about a year.
const historyDays = 365

// pricesBeforeGoLive is how many days before go-live the price paths start,
// so that the first day valued has a close.
const pricesBeforeGoLive = 7

// The streams of the seeded generator, one per use, so that adding draws to
// one doesn't change the others.
const (
	streamIBKR uint64 = 0x1b0001
	streamXTB  uint64 = 0x470002
	streamWalk uint64 = 0x9a0000
)

// Dataset is the demo dataset: a synthetic portfolio, its history from
// go-live to today, and the valuation inputs it needs besides FX rates.
type Dataset struct {
	Seed   uint64
	Today  kernel.Date
	GoLive kernel.Date

	Sleeves     []portfolio.Sleeve
	Accounts    []ledger.Account
	Instruments []Instrument
	Bonds       []BondSeries
	// TaxWrapper is the bond account's: IKE.
	TaxWrapper pl.TaxWrapper

	// IBKROpening is the Flex statement the IBKR account is seeded from,
	// and IBKRActivity its monthly activity statements, oldest first.
	IBKROpening  []byte
	IBKRActivity [][]byte

	XTBOpening    Batch
	XTBActivity   Batch
	BondsOpening  Batch
	BondsActivity Batch

	Prices []valuation.Price
	CPI    []valuation.IndexReading
	// CrossCheckPrices are IBKR's synthetic mark prices at each monthly
	// statement's end, one of them beyond the price check threshold.
	CrossCheckPrices []valuation.CrossCheckPrice
}

// GoLive returns the demo's go-live date for today: a year before.
func GoLive(today kernel.Date) kernel.Date { return today.AddDays(-historyDays) }

// Generate returns the demo dataset for seed as of today. It depends on
// nothing else: the same seed and today always give the same dataset.
func Generate(seed uint64, today kernel.Date) (Dataset, error) {
	goLive := GoLive(today)
	ds := Dataset{Seed: seed, Today: today, GoLive: goLive, Sleeves: Sleeves(), Accounts: Accounts(), TaxWrapper: pl.IKE, CPI: CPIReadings()}
	m, err := newMarket(seed, goLive.AddDays(-pricesBeforeGoLive), today, &ds)
	if err != nil {
		return Dataset{}, err
	}
	if ds.Bonds, err = Bonds(); err != nil {
		return Dataset{}, err
	}
	h := simulateIBKR(m, newRNG(seed^streamIBKR), goLive, today)
	ds.CrossCheckPrices = crossCheckPrices(m, &h)
	ds.IBKROpening = openingFlex(&h)
	for i := range h.statements {
		ds.IBKRActivity = append(ds.IBKRActivity, activityFlex(&h.statements[i]))
	}
	xtb, err := simulateXTB(m, newRNG(seed^streamXTB), goLive, today)
	if err != nil {
		return Dataset{}, err
	}
	bonds, err := simulateBonds(goLive, today, ds.CPI)
	if err != nil {
		return Dataset{}, err
	}
	ds.XTBOpening, ds.XTBActivity = xtb.opening, xtb.activity
	ds.BondsOpening, ds.BondsActivity = bonds.opening, bonds.activity
	return ds, nil
}

// newMarket returns the demo's instruments with their price paths from
// first to last, and stores them and their prices in ds.
func newMarket(seed uint64, first, last kernel.Date, ds *Dataset) (*market, error) {
	ins, err := Instruments()
	if err != nil {
		return nil, err
	}
	usd, err1 := kernel.ParseCurrency(codeUSD)
	eur, err2 := kernel.ParseCurrency(codeEUR)
	if err := firstError(err1, err2); err != nil {
		return nil, fmt.Errorf("demo market: %w", err)
	}
	m := &market{
		bySymbol: map[string]*Instrument{}, paths: map[*Instrument]pricePath{},
		usd: usd, eur: eur, pln: pl.Nominal().Currency(),
	}
	ds.Instruments = ins
	for i := range ds.Instruments {
		in := &ds.Instruments[i]
		path := walk(in, newRNG(seed^streamWalk^uint64(i+1)), first, last)
		m.bySymbol[in.Instrument.Symbol], m.paths[in] = in, path
		ds.Prices = append(ds.Prices, prices(in, path, first, last)...)
	}
	return m, nil
}
