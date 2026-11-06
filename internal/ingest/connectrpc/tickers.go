package connectrpc

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
)

// tickersFrom adds to m the tickers d names instruments by and their
// entries (D-116), those d names by broker account to m.ByBrokerAccount.
// Every ticker must be valid, and every entry must name one of them and
// book a valid transaction.
func tickersFrom(d *connectorv1.Delivery, m *ingest.Mapping) error {
	known := map[string]bool{}
	for i, t := range d.GetTickers() {
		tk := ingest.Ticker{Symbol: t.GetSymbol(), Code: t.GetCode(), Country: t.GetCountry(), Name: t.GetName()}
		if err := tk.Validate(); err != nil {
			return invalid(fmt.Sprintf("tickers[%d]", i), err)
		}
		if !known[tk.Symbol] {
			known[tk.Symbol] = true
			m.Tickers = append(m.Tickers, tk)
		}
	}
	for i, t := range d.GetTickerTransactions() {
		field := fmt.Sprintf("ticker_transactions[%d]", i)
		e, err := tickerEntryFrom(field, t)
		if err != nil {
			return err
		}
		if !known[e.Symbol] {
			return invalid(field+".symbol names no ticker of the delivery", nil)
		}
		id, err := owner(field, t.GetAccountId(), t.GetBrokerAccount())
		if err != nil {
			return err
		}
		if id == "" {
			m.TickerEntries = append(m.TickerEntries, e)
			continue
		}
		brokerEntriesOf(m, id).TickerEntries = append(brokerEntriesOf(m, id).TickerEntries, e)
	}
	return tickerAssertionsFrom(d, m, known)
}

// tickerAssertionsFrom adds to m the ticker assertions of d (D-155), those
// d names by broker account to m.ByBrokerAccount. Each must name one of
// the delivery's tickers, a date and a quantity that is not negative.
func tickerAssertionsFrom(d *connectorv1.Delivery, m *ingest.Mapping, known map[string]bool) error {
	for i, t := range d.GetTickerAssertions() {
		field := fmt.Sprintf("ticker_assertions[%d]", i)
		if !known[t.GetSymbol()] {
			return invalid(field+".symbol names no ticker of the delivery", nil)
		}
		on, err := dateFrom(field+".date", t.GetDate())
		if err != nil {
			return err
		}
		qty, err := decimalFrom(field+".quantity", t.GetQuantity())
		if err != nil {
			return err
		}
		a := ingest.TickerAssertion{
			Symbol: t.GetSymbol(), Account: ledger.AccountID(t.GetAccountId()), Date: on, Quantity: kernel.NewQuantity(qty), Ref: t.GetRef(),
		}
		if vErr := a.Validate(); vErr != nil {
			return invalid(field, vErr)
		}
		id, err := owner(field, t.GetAccountId(), t.GetBrokerAccount())
		if err != nil {
			return err
		}
		if id == "" {
			m.TickerAssertions = append(m.TickerAssertions, a)
			continue
		}
		a.Account = ""
		brokerEntriesOf(m, id).TickerAssertions = append(brokerEntriesOf(m, id).TickerAssertions, a)
	}
	return nil
}

// brokerEntriesOf returns m's entries of broker account id, adding them.
func brokerEntriesOf(m *ingest.Mapping, id ingest.BrokerAccountID) *ingest.BrokerEntries {
	for i := range m.ByBrokerAccount {
		if m.ByBrokerAccount[i].Account == id {
			return &m.ByBrokerAccount[i]
		}
	}
	m.ByBrokerAccount = append(m.ByBrokerAccount, ingest.BrokerEntries{Account: id})
	return &m.ByBrokerAccount[len(m.ByBrokerAccount)-1]
}

// tickerEntryFrom translates one ticker transaction, which must book a
// valid transaction: a buy, sell, dividend or opening lot (D-136).
func tickerEntryFrom(field string, t *connectorv1.TickerTransaction) (ingest.TickerEntry, error) {
	kind, err := transactionKindFrom(field, t.GetKind())
	if err != nil {
		return ingest.TickerEntry{}, err
	}
	date, err := dateFrom(field+".date", t.GetDate())
	if err != nil {
		return ingest.TickerEntry{}, err
	}
	var amount kernel.Money
	if t.GetAmount() != nil || t.GetOpenPrice() == "" {
		if amount, err = moneyFrom(field+".amount", t.GetAmount()); err != nil {
			return ingest.TickerEntry{}, err
		}
	}
	e := ingest.TickerEntry{
		Symbol: t.GetSymbol(),
		Header: ledger.TransactionHeader{
			Key:     ledger.TransactionKey{Source: t.GetKey().GetSource(), ExternalID: t.GetKey().GetExternalId()},
			Account: ledger.AccountID(t.GetAccountId()), Date: date,
		},
		Kind: kind, Amount: amount, Description: t.GetDescription(), Ref: t.GetRef(),
	}
	if t.GetCharge() != nil {
		if e.Charge, err = moneyFrom(field+".charge", t.GetCharge()); err != nil {
			return ingest.TickerEntry{}, err
		}
	}
	if err := optionalFrom(field, t, &e); err != nil {
		return ingest.TickerEntry{}, err
	}
	if err := e.Validate(); err != nil {
		return ingest.TickerEntry{}, invalid(field, err)
	}
	return e, nil
}

// optionalFrom reads into e the fields a ticker transaction may leave
// empty: an opening lot's lot_opened and open_price (D-159), and the
// quantity.
func optionalFrom(field string, t *connectorv1.TickerTransaction, e *ingest.TickerEntry) error {
	if t.GetOpenPrice() != "" {
		price, err := decimalFrom(field+".open_price", t.GetOpenPrice())
		if err != nil {
			return err
		}
		e.OpenPrice = price
	}
	if t.GetLotOpened() != "" {
		opened, err := dateFrom(field+".lot_opened", t.GetLotOpened())
		if err != nil {
			return err
		}
		e.Opened = opened
	}
	if t.GetQuantity() != "" {
		q, err := decimalFrom(field+".quantity", t.GetQuantity())
		if err != nil {
			return err
		}
		e.Quantity = kernel.NewQuantity(q)
	}
	return nil
}

// tickersTo writes tickers for a lookup request.
func tickersTo(tickers []ingest.Ticker) []*connectorv1.Ticker {
	out := make([]*connectorv1.Ticker, len(tickers))
	for i := range tickers {
		t := &tickers[i]
		out[i] = &connectorv1.Ticker{Symbol: t.Symbol, Code: t.Code, Country: t.Country, Name: t.Name}
	}
	return out
}

// proposalsFrom translates a lookup's proposals: each must name an asked
// symbol and a valid ISIN, and a symbol proposed twice keeps its first.
func proposalsFrom(asked []ingest.Ticker, ps []*connectorv1.InstrumentProposal) ([]ingest.ISINProposal, error) {
	want := map[string]bool{}
	for i := range asked {
		want[asked[i].Symbol] = true
	}
	var out []ingest.ISINProposal
	seen := map[string]bool{}
	for i, p := range ps {
		field := fmt.Sprintf("proposals[%d]", i)
		if !want[p.GetSymbol()] {
			return nil, invalid(field+".symbol was not asked for", nil)
		}
		isin, err := ledger.ParseISIN(p.GetIsin())
		if err != nil {
			return nil, invalid(field+".isin", err)
		}
		if seen[p.GetSymbol()] {
			continue
		}
		seen[p.GetSymbol()] = true
		name := []rune(p.GetName())
		if len(name) > maxProposalName {
			name = name[:maxProposalName]
		}
		currency := p.GetCurrency()
		if _, _, ok := ingest.ParseTradingCurrency(currency); !ok {
			currency = ""
		}
		out = append(out, ingest.ISINProposal{Symbol: p.GetSymbol(), ISIN: isin, Name: string(name), Currency: currency})
	}
	return out, nil
}

// maxProposalName bounds the name kept with a proposal.
const maxProposalName = 200

// LookupInstruments asks instance, a market-data connector, for the ISIN of
// each of tickers (D-116). A connector that can't look instruments up,
// answering UNIMPLEMENTED, wraps app.ErrNoEntries; one that can't be
// reached app.ErrConnectorUnavailable.
func (c *Client) LookupInstruments(ctx context.Context, instance string, tickers []ingest.Ticker) ([]ingest.ISINProposal, error) {
	call, err := c.service(ctx, instance)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", app.ErrConnectorUnavailable, err)
	}
	req := connect.NewRequest(&connectorv1.LookupInstrumentsRequest{Tickers: tickersTo(tickers)})
	req.Header().Set("Authorization", call.auth)
	resp, err := call.svc.LookupInstruments(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("connector %s: %w", instance, entryError(err))
	}
	out, err := proposalsFrom(tickers, resp.Msg.GetProposals())
	if err != nil {
		return nil, fmt.Errorf("connector %s: %w", instance, err)
	}
	return out, nil
}
