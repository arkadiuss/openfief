package ui

import (
	"encoding/json"
	"time"

	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// offlineDB names the IndexedDB database in which the installed app keeps
// the last overview, encrypted, for offline reading (D-151). offline.js
// writes it and the service worker deletes it on sign-out; both spell it
// this way.
const offlineDB = "openfief-offline"

// offlineVersion is the shape of the kept overview; offline.js shows only
// a copy of the version it knows.
const offlineVersion = 1

// offlineOverview is everything of the overview the installed app keeps
// on the device for reading offline (D-151): the total's three values,
// each account's three values with the names that tell the accounts apart,
// and the allocation's shares, written as the page writes them, with the
// date they are valued on and when the session ends. It holds no account
// ID, broker account number, transaction, lot, holding or amount of an
// allocation group: nothing else of the record is kept.
type offlineOverview struct {
	Version int    `json:"v"`
	Date    string `json:"date"`
	// SessionEnds is when the session that rendered the overview ends,
	// RFC 3339 in UTC; past it, offline.js wipes the copy unread.
	SessionEnds string            `json:"sessionEnds"`
	Total       offlineValues     `json:"total"`
	Accounts    []offlineAccount  `json:"accounts"`
	Allocation  offlineAllocation `json:"allocation"`
}

// offlineValues are the gross, net and liquidation values as the page
// writes them, an empty one missing.
type offlineValues struct {
	Gross       string `json:"gross"`
	Net         string `json:"net"`
	Liquidation string `json:"liquidation"`
}

// offlineAccount is one account of the kept overview: the owner's name for
// it, its institution and sleeve, and its three values; Valued is false
// when nothing of it is valued on the date.
type offlineAccount struct {
	Name        string        `json:"name"`
	Institution string        `json:"institution"`
	Sleeve      string        `json:"sleeve"`
	Valued      bool          `json:"valued"`
	Values      offlineValues `json:"values"`
}

// offlineAllocation is the allocation's split as shares only: the
// dimension, the value it splits, and each group's share of the total.
type offlineAllocation struct {
	By     string         `json:"by"`
	Value  string         `json:"value"`
	Shares []offlineShare `json:"shares"`
}

// offlineShare is one allocation group and its share, such as "42.5%".
type offlineShare struct {
	Key   string `json:"key"`
	Share string `json:"share"`
}

// threeValues writes a date's total as the page does, amounts in f.
func threeValues(t *viewsv1.DateTotal, f numberFormat) offlineValues {
	return offlineValues{
		Gross: f.money(t.GetValueReporting()), Net: f.money(t.GetNetReporting()),
		Liquidation: f.money(t.GetLiquidationReporting()),
	}
}

// buildOfflineOverview picks from the overview d what the installed app
// may keep (D-151), amounts written in f, with ends, when the session
// ends. Without a total on the date the gross value is the allocation's
// total. An account's sleeve is written by its name, as the overview
// groups it.
func buildOfflineOverview(d *overviewData, ends time.Time, f numberFormat) offlineOverview {
	o := offlineOverview{Version: offlineVersion, Date: d.Date}
	if !ends.IsZero() {
		o.SessionEnds = ends.UTC().Format(time.RFC3339)
	}
	var alloc *allocationView
	if d.Over != nil {
		alloc = d.Over.Allocation
	}
	if d.Total != nil {
		o.Total = threeValues(d.Total, f)
	} else if alloc != nil {
		o.Total.Gross = f.money(alloc.Total)
	}
	o.Accounts = offlineAccounts(d, f)
	o.Allocation.Shares = []offlineShare{}
	if alloc != nil {
		o.Allocation.By, o.Allocation.Value = alloc.Dimension.Label, alloc.Value.Label
		for _, g := range alloc.Groups {
			o.Allocation.Shares = append(o.Allocation.Shares, offlineShare{Key: g.Key, Share: g.Share})
		}
	}
	return o
}

// offlineAccounts returns each account of the overview d as the installed
// app keeps it (D-151): its names and three values, amounts written in f,
// and its sleeve by its name from the overview's sleeve groups, its ID
// without one.
func offlineAccounts(d *overviewData, f numberFormat) []offlineAccount {
	sleeves := map[string]string{}
	for _, g := range d.Sleeves {
		sleeves[g.ID] = g.Name
	}
	out := make([]offlineAccount, 0, len(d.Accounts))
	for _, a := range d.Accounts {
		sleeve := a.Account.GetSleeve()
		if name := sleeves[sleeve]; name != "" {
			sleeve = name
		}
		row := offlineAccount{
			Name: a.Account.GetName(), Institution: a.Account.GetInstitution(), Sleeve: sleeve,
			Valued: a.Total != nil,
		}
		if a.Total != nil {
			row.Values = threeValues(a.Total, f)
		}
		out = append(out, row)
	}
	return out
}

// encode returns the kept overview as JSON for the overview page's data
// attribute, which offline.js encrypts and stores.
func (o *offlineOverview) encode() string {
	b, err := json.Marshal(o)
	if err != nil { // strings, ints and bools only: it always encodes
		return ""
	}
	return string(b)
}
