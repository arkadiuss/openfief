package views

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/valuation"
)

// ErrUnknownDimension reports a Dimension outside the declared constants, or
// text that names none of them.
var ErrUnknownDimension = errors.New("unknown allocation dimension")

// polishISINPrefix is the country code that Polish retail treasury bond ISINs
// start with.
const polishISINPrefix = "PL"

// Dimension is what an allocation groups the portfolio by (D-30).
type Dimension int

// The dimensions. The zero Dimension is not a dimension.
const (
	// ByAssetClass groups by AssetClass.
	ByAssetClass Dimension = iota + 1
	// ByCurrency groups by the currency a value is in before conversion.
	ByCurrency
	// ByInstitution groups by the institution that holds the account.
	ByInstitution
	// BySleeve groups by the account's sleeve.
	BySleeve
)

// String returns the dimension's name: asset_class, currency, institution or
// sleeve.
func (d Dimension) String() string {
	switch d {
	case ByAssetClass:
		return "asset_class"
	case ByCurrency:
		return "currency"
	case ByInstitution:
		return "institution"
	case BySleeve:
		return "sleeve"
	}
	return fmt.Sprintf("Dimension(%d)", int(d))
}

// ParseDimension parses a dimension name as String writes it.
func ParseDimension(s string) (Dimension, error) {
	for _, d := range []Dimension{ByAssetClass, ByCurrency, ByInstitution, BySleeve} {
		if d.String() == s {
			return d, nil
		}
	}
	return 0, fmt.Errorf("%w: %q", ErrUnknownDimension, s)
}

// AssetClass is the kind of asset an allocation item is, for the allocation
// chart (D-30).
type AssetClass int

// The asset classes. The zero AssetClass is not a class.
const (
	// AssetCash is cash in an account.
	AssetCash AssetClass = iota + 1
	// AssetTreasuryBond is a Polish retail treasury bond series.
	AssetTreasuryBond
	// AssetListed is any other instrument: in M0, a stock or an ETF.
	AssetListed
)

// String returns the class's name: cash, treasury_bond or listed.
func (c AssetClass) String() string {
	switch c {
	case AssetCash:
		return "cash"
	case AssetTreasuryBond:
		return "treasury_bond"
	case AssetListed:
		return "listed"
	}
	return fmt.Sprintf("AssetClass(%d)", int(c))
}

// AssetClassOf returns the asset class of a holding of in: a treasury bond
// when its ISIN is Polish and its venue symbol is a bond series code, such as
// EDO0536, and listed otherwise.
func AssetClassOf(in ledger.Instrument) AssetClass {
	if _, err := pl.ParseSeries(in.Symbol); err == nil && strings.HasPrefix(in.ISIN.String(), polishISINPrefix) {
		return AssetTreasuryBond
	}
	return AssetListed
}

// AllocationItem is one valued holding or cash balance of an account on a
// date: what an allocation groups.
type AllocationItem struct {
	Account    ledger.Account
	AssetClass AssetClass
	// Value is in its original currency.
	Value kernel.Money
	// PLN is Value converted to PLN.
	PLN kernel.Money
}

// AllocationGroup is the PLN value of the items that share one key: an asset
// class, a currency code, an institution or a sleeve ID.
type AllocationGroup struct {
	Key string
	PLN kernel.Money
}

// Allocation is the portfolio's value in PLN split by one dimension.
type Allocation struct {
	By     Dimension
	Total  kernel.Money
	Groups []AllocationGroup
}

// Allocate groups items by d and sums each group's PLN value: largest group
// first, then by key. Total is the sum of all groups, zero PLN when there are
// no items. A PLN value in another currency is an error.
func Allocate(items []AllocationItem, d Dimension) (Allocation, error) {
	sums := map[string]kernel.Money{}
	total := kernel.NewMoney(kernel.Decimal{}, valuation.PLN())
	for i := range items {
		key, err := groupKey(&items[i], d)
		if err != nil {
			return Allocation{}, err
		}
		if sums[key], err = addPLN(sums[key], items[i].PLN); err != nil {
			return Allocation{}, fmt.Errorf("allocation by %s, %s: %w", d, key, err)
		}
		if total, err = addPLN(total, items[i].PLN); err != nil {
			return Allocation{}, fmt.Errorf("allocation by %s: %w", d, err)
		}
	}
	groups := make([]AllocationGroup, 0, len(sums))
	for k, v := range sums {
		groups = append(groups, AllocationGroup{Key: k, PLN: v})
	}
	slices.SortFunc(groups, func(a, b AllocationGroup) int {
		if c := b.PLN.Amount().Cmp(a.PLN.Amount()); c != 0 {
			return c
		}
		return cmp.Compare(a.Key, b.Key)
	})
	return Allocation{By: d, Total: total, Groups: groups}, nil
}

// groupKey returns the key item is grouped under by d.
func groupKey(item *AllocationItem, d Dimension) (string, error) {
	switch d {
	case ByAssetClass:
		return item.AssetClass.String(), nil
	case ByCurrency:
		return item.Value.Currency().String(), nil
	case ByInstitution:
		return item.Account.Institution, nil
	case BySleeve:
		return string(item.Account.Sleeve), nil
	}
	return "", fmt.Errorf("%w: %s", ErrUnknownDimension, d)
}

// ValueHoldings pairs account's holdings with their valuations dated on: a
// holding with a valuation of its instrument in the account on that date
// becomes an allocation item, and one without is returned as unvalued.
func ValueHoldings(account ledger.Account, holdings []ledger.Holding, vals []HoldingValuation, on kernel.Date) ([]AllocationItem, []ledger.Holding) {
	var items []AllocationItem
	var unvalued []ledger.Holding
	for i := range holdings {
		h := &holdings[i]
		v, ok := valuationOf(vals, account.ID, h.Instrument, on)
		if !ok {
			unvalued = append(unvalued, *h)
			continue
		}
		items = append(items, AllocationItem{Account: account, AssetClass: AssetClassOf(h.Instrument), Value: v.Value, PLN: v.PLN})
	}
	return items, unvalued
}

// valuationOf returns the valuation in vals of in, held in account, dated on.
func valuationOf(vals []HoldingValuation, account ledger.AccountID, in ledger.Instrument, on kernel.Date) (HoldingValuation, bool) {
	for i := range vals {
		if v := &vals[i]; v.Account == account && v.Instrument == in && v.Date.Equal(on) {
			return *v, true
		}
	}
	return HoldingValuation{}, false
}
