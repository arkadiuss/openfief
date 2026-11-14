package main

import (
	"strings"
	"testing"

	"github.com/arkadiuss/openfief/internal/kernel"
)

func buildPortfolio(t *testing.T) *Portfolio {
	t.Helper()
	seed, err := BuildSeed(false)
	if err != nil {
		t.Fatalf("BuildSeed: %v", err)
	}
	pf, err := NewPortfolio(seed)
	if err != nil {
		t.Fatalf("NewPortfolio: %v", err)
	}
	return pf
}

func TestPortfolioIsValuedOnEveryDayAndMovesWithPrices(t *testing.T) {
	t.Parallel()
	pf := buildPortfolio(t)
	first, last := pf.Days[0], pf.Days[len(pf.Days)-1]
	if got, want := last.Date.String(), benchDay; got != want {
		t.Fatalf("last day = %s, want %s", got, want)
	}
	if first.Gross().Sign() <= 0 || last.Gross().Sign() <= 0 {
		t.Fatalf("gross values %s and %s, want positive ones", first.Gross(), last.Gross())
	}
	if len(pf.Days) != first.Date.DaysUntil(last.Date)+1 {
		t.Errorf("%d days between %s and %s", len(pf.Days), first.Date, last.Date)
	}
	if first.Index.Cmp(kernel.NewDecimal(indexStart, 0)) != 0 {
		t.Errorf("the index starts at %s, want %d", first.Index, indexStart)
	}
	for acc, byCur := range last.CashByCurrency {
		for cur, amount := range byCur {
			if amount.Sign() < 0 {
				t.Errorf("%s holds %s %s: replaying the ledger went wrong", acc, amount, cur)
			}
		}
	}
}

func TestPortfolioIsDeterministic(t *testing.T) {
	t.Parallel()
	a, b := buildPortfolio(t), buildPortfolio(t)
	if len(a.Days) != len(b.Days) {
		t.Fatalf("%d and %d days", len(a.Days), len(b.Days))
	}
	for i := range a.Days {
		if a.Days[i].String() != b.Days[i].String() || !a.Days[i].Index.Equal(b.Days[i].Index) {
			t.Fatalf("day %d differs: %s and %s", i, a.Days[i].String(), b.Days[i].String())
		}
	}
}

func TestTheSeedDropsInAMonth(t *testing.T) {
	t.Parallel()
	month, drop, ok := buildPortfolio(t).DropMonth()
	if !ok || drop.Sign() >= 0 {
		t.Fatalf("DropMonth = %q, %s, %v: the bench's A19 needs a month the return fell in", month, drop, ok)
	}
}

func TestVariantPlantsOnlyThePayloads(t *testing.T) {
	t.Parallel()
	clean, err := BuildSeed(false)
	if err != nil {
		t.Fatal(err)
	}
	variant, err := BuildSeed(true)
	if err != nil {
		t.Fatal(err)
	}
	if n := changedComments(t, clean, variant); n != 1 {
		t.Errorf("variant changed %d comments, want one", n)
	}
	if n := changedNames(t, clean, variant); n != 1 {
		t.Errorf("variant changed %d names, want one", n)
	}
	if n := changedAccountNames(t, clean, variant); n != 1 {
		t.Errorf("variant changed %d account names, want one", n)
	}
	for _, in := range clean.Instruments {
		for _, p := range payloads() {
			if strings.Contains(in.Name, p) {
				t.Errorf("the clean seed names %q", in.Name)
			}
		}
	}
}

// changedComments counts the transactions whose comment the variant changed,
// and checks each is the planted one.
func changedComments(t *testing.T, clean, variant *Seed) int {
	t.Helper()
	n := 0
	for i := range variant.Transactions {
		if clean.Transactions[i].Description == variant.Transactions[i].Description {
			continue
		}
		n++
		if variant.Transactions[i].Description != payloadComment {
			t.Errorf("transaction %d comment = %q", i, variant.Transactions[i].Description)
		}
	}
	return n
}

// changedNames counts the instruments whose name the variant changed, and
// checks each carries the planted one.
func changedNames(t *testing.T, clean, variant *Seed) int {
	t.Helper()
	n := 0
	for isin, in := range variant.Instruments {
		if clean.Instruments[isin].Name == in.Name {
			continue
		}
		n++
		if !strings.Contains(in.Name, payloadName) {
			t.Errorf("instrument name = %q", in.Name)
		}
	}
	return n
}

// changedAccountNames counts the accounts whose name the variant changed, and
// checks each carries the planted comment.
func changedAccountNames(t *testing.T, clean, variant *Seed) int {
	t.Helper()
	n := 0
	for i := range variant.Dataset.Accounts {
		name := variant.Dataset.Accounts[i].Name
		if clean.Dataset.Accounts[i].Name == name {
			continue
		}
		n++
		if !strings.Contains(name, payloadComment) {
			t.Errorf("account name = %q", name)
		}
	}
	return n
}
