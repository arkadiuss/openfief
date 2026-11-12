// Package ghostfoliocheck reads, for the end-to-end suite's check of D-200's
// criterion 9, what the Ghostfolio view plugin reports and what a real
// Ghostfolio holds after it synced: each pass of the reconcile loop the
// plugin logged, with its counts or its error; the transaction key an
// activity's comment names (D-4); the units of a bond series an owned
// account holds in Ghostfolio, to compare with the read API's holding; an
// owned account's cash balance on a date (D-108), to compare with the read
// API's cash; a line per activity and balance, IDs included, to tell that a
// sync changed nothing; and the body of the hand edit an owner makes in
// Ghostfolio's UI, which the next sync must revert. It only reads log bytes
// and its own copies of the activities and balances the view's client
// returns, since only cmd imports an adapter; the suite (test/e2e,
// ghostfolio_test.go) runs Ghostfolio and the plugin and converts them. It
// is never released.
package ghostfoliocheck

import (
	"bufio"
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/arkadiuss/openfief/internal/kernel"
)

// CommentPrefix starts the comment of every activity the view writes, as
// internal/views/ghostfolio's CommentPrefix does; the transaction's key
// follows it (D-4).
const CommentPrefix = "openfief:"

// The activity types the view writes (internal/views/ghostfolio).
const (
	TypeBuy      = "BUY"
	TypeSell     = "SELL"
	TypeDividend = "DIVIDEND"
	TypeInterest = "INTEREST"
	TypeFee      = "FEE"
)

// Activity is a Ghostfolio activity as the view's client returns it, with
// Ghostfolio's ID.
type Activity struct {
	ID, AccountID, Comment, Type string
	Date                         kernel.Date
	Currency                     kernel.Currency
	DataSource, Symbol           string
	Quantity, UnitPrice, Fee     kernel.Decimal
}

// Balance is an account's cash balance at the end of a date, as the view's
// client returns it, with Ghostfolio's ID.
type Balance struct {
	ID, AccountID string
	Date          kernel.Date
	Value         kernel.Decimal
}

// Report holds the counts of a finished pass, as the view's report has them.
type Report struct {
	AccountsCreated, Imported, Deleted, AccountsCleared, BalancesSet, BalancesDeleted int
}

// The messages the plugin logs after each pass (plugins/ghostfolio, loop).
const (
	syncedMsg = "synced"
	failedMsg = "sync failed"
)

// feeMark starts the suffix of a fee split off its activity: "#fee:" and
// the fee's currency (internal/views/ghostfolio, Project).
const feeMark = "#fee:"

// logLineBytes bounds one line of the plugin's log.
const logLineBytes = 1 << 20

// errBadLog reports a log line of a pass that isn't the plugin's JSON.
var errBadLog = errors.New("unreadable sync log line")

// Sync is one pass of the view's reconcile loop as the plugin logged it:
// the counts of a pass that finished, or the error of one that failed.
type Sync struct {
	// Report holds the counts.
	Report Report
	// LeftOut is how many transactions the pass left out.
	LeftOut int
	// Err is a failed pass's error; empty for one that finished.
	Err string
}

// countKeys are the keys of a finished pass's counts in the plugin's log,
// in the order passOf reads them.
func countKeys() []string {
	return []string{
		"accounts_created", "imported", "deleted", "accounts_cleared", "balances_set", "balances_deleted", "left_out",
	}
}

// Syncs returns the passes the plugin logged, in order. Lines of anything
// else, JSON or not, are skipped; a line that names a pass but isn't the
// plugin's JSON fails.
func Syncs(log []byte) ([]Sync, error) {
	var out []Sync
	sc := bufio.NewScanner(bytes.NewReader(log))
	sc.Buffer(make([]byte, 0, logLineBytes), logLineBytes)
	for sc.Scan() {
		line := sc.Bytes()
		if !bytes.Contains(line, []byte(`"`+syncedMsg+`"`)) && !bytes.Contains(line, []byte(`"`+failedMsg+`"`)) {
			continue
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(line, &fields); err != nil {
			return nil, fmt.Errorf("%w: %w", errBadLog, err)
		}
		pass, ok, err := passOf(fields)
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, pass)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("reading the log: %w", err)
	}
	return out, nil
}

// passOf reads a pass from the fields of a line of the plugin's JSON log,
// and reports whether the line is one: "synced" with every count, or
// "sync failed" with its error.
func passOf(fields map[string]json.RawMessage) (Sync, bool, error) {
	raw, ok := fields["msg"]
	if !ok {
		return Sync{}, false, nil
	}
	var msg string
	if err := json.Unmarshal(raw, &msg); err != nil {
		return Sync{}, false, fmt.Errorf("%w: msg: %w", errBadLog, err)
	}
	switch msg {
	case failedMsg:
		var e string
		if err := json.Unmarshal(fields["error"], &e); err != nil {
			return Sync{}, false, fmt.Errorf("%w: error: %w", errBadLog, err)
		}
		return Sync{Err: e}, true, nil
	case syncedMsg:
		keys := countKeys()
		counts := make([]int, len(keys))
		for i, k := range keys {
			if err := json.Unmarshal(fields[k], &counts[i]); err != nil {
				return Sync{}, false, fmt.Errorf("%w: %s: %w", errBadLog, k, err)
			}
		}
		return Sync{LeftOut: counts[6], Report: Report{
			AccountsCreated: counts[0], Imported: counts[1], Deleted: counts[2],
			AccountsCleared: counts[3], BalancesSet: counts[4], BalancesDeleted: counts[5],
		}}, true, nil
	}
	return Sync{}, false, nil
}

// CommentKey returns the transaction key an activity's comment names, and
// whether the view wrote the comment: CommentPrefix, the key, and for a fee
// split off its activity "#fee:" and a currency.
func CommentKey(comment string) (string, bool) {
	key, ok := strings.CutPrefix(comment, CommentPrefix)
	if !ok || key == "" {
		return "", false
	}
	if i := strings.LastIndex(key, feeMark); i >= 0 {
		if _, err := kernel.ParseCurrency(key[i+len(feeMark):]); err == nil {
			key = key[:i]
		}
	}
	return key, key != ""
}

// NetUnits returns the units of symbol that the activities of account
// hold: what its BUYs bought less what its SELLs sold.
func NetUnits(acts []Activity, account, symbol string) kernel.Decimal {
	var units kernel.Decimal
	for i := range acts {
		a := &acts[i]
		if a.AccountID != account || a.Symbol != symbol {
			continue
		}
		switch a.Type {
		case TypeBuy:
			units = units.Add(a.Quantity)
		case TypeSell:
			units = units.Sub(a.Quantity)
		case TypeDividend, TypeInterest, TypeFee:
		}
	}
	return units
}

// Latest returns account's cash balance at the end of date: the value of
// its latest balance on or before date, zero when it has none, as
// Ghostfolio reads an account's cash.
func Latest(bs []Balance, account string, date kernel.Date) kernel.Decimal {
	var (
		value kernel.Decimal
		at    kernel.Date
	)
	for i := range bs {
		b := &bs[i]
		if b.AccountID != account || b.Date.After(date) || (!at.IsZero() && !b.Date.After(at)) {
			continue
		}
		value, at = b.Value, b.Date
	}
	return value
}

// Describe returns a line per activity, with Ghostfolio's ID and every
// field the view writes, in order: two syncs with the same lines left
// every activity as it was, none imported, deleted or edited.
func Describe(acts []Activity) []string {
	out := make([]string, 0, len(acts))
	for i := range acts {
		a := &acts[i]
		out = append(out, strings.Join([]string{
			a.ID, a.AccountID, a.Comment, a.Type, a.Date.String(), a.Currency.String(), a.DataSource, a.Symbol,
			a.Quantity.String(), a.UnitPrice.String(), a.Fee.String(),
		}, " "))
	}
	slices.Sort(out)
	return out
}

// DescribeBalances returns a line per cash balance, with Ghostfolio's ID,
// in order, as Describe does for activities.
func DescribeBalances(bs []Balance) []string {
	out := make([]string, 0, len(bs))
	for i := range bs {
		b := &bs[i]
		out = append(out, strings.Join([]string{b.ID, b.AccountID, b.Date.String(), b.Value.String()}, " "))
	}
	slices.Sort(out)
	return out
}

// Editable returns the activity of acts an owner edits by hand in the
// check: the first by ID that holds units, a BUY, SELL, DIVIDEND or
// INTEREST, whose quantity one more unit changes; false when there is none.
func Editable(acts []Activity) (Activity, bool) {
	sorted := slices.SortedFunc(slices.Values(acts), func(a, b Activity) int { return cmp.Compare(a.ID, b.ID) })
	for i := range sorted {
		if sorted[i].Type != TypeFee && sorted[i].Quantity.Sign() > 0 {
			return sorted[i], true
		}
	}
	return Activity{}, false
}

// Edited returns the body of PUT /api/v1/activities/{id} that edits a as
// the owner would in Ghostfolio's UI: one more unit, every other field,
// the comment among them, kept. A manual edit keeps its comment, so only
// comparing fields finds it (D-64).
func Edited(a *Activity) map[string]any {
	return map[string]any{
		"accountId": a.AccountID, "comment": a.Comment, "currency": a.Currency.String(), "dataSource": a.DataSource,
		"date": a.Date.String() + "T00:00:00.000Z", "fee": json.Number(a.Fee.String()), "id": a.ID,
		"quantity": json.Number(a.Quantity.Add(kernel.NewDecimal(1, 0)).String()), "symbol": a.Symbol,
		"type": a.Type, "unitPrice": json.Number(a.UnitPrice.String()),
	}
}
